package pomodoro

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type PomodoroStatus int

const (
	PomodoroStatusStop PomodoroStatus = iota
	PomodoroStatusTask
	PomodoroStatusBreakTime
)

const (
	PomodoroTaskDuration            = "25m"
	PomodoroBreakDuration           = "5m"
	PomodoroWarningEndBreakDuration = "10s"
)

// Start() でスタート
// Stop() でストップ
// Stop() 後に struct を破棄する
type Pomodoro struct {
	session       *discordgo.Session
	guildID       ChannelID
	textChannelID ChannelID
	// Joining users
	members                 map[UserID]discordgo.User
	status                  PomodoroStatus `default:"PomodoroStatusStop"`
	taskDuration            time.Duration
	breakDuration           time.Duration
	warningEndBreakDuration time.Duration
	timer                   *time.Timer
	taskEndTimerCh          chan struct{}
	breakEndTimerCh         chan struct{}
	stopCh                  chan struct{}
	wg                      sync.WaitGroup
}

func NewPomodoro(session *discordgo.Session, guildID ChannelID, textChannelID ChannelID) (*Pomodoro, error) {

	taskDuration, err := time.ParseDuration(PomodoroTaskDuration)
	if err != nil {
		fmt.Errorf("Error parsing PomodoroTaskDuration (%v): %v", PomodoroTaskDuration, err)
		return nil, err
	}

	breakDuration, err := time.ParseDuration(PomodoroBreakDuration)
	if err != nil {
		fmt.Errorf("Error parsing PomodoroBreakDuration (%v): %v", PomodoroBreakDuration, err)
		return nil, err
	}

	warningEndBreakDuration, err := time.ParseDuration(PomodoroWarningEndBreakDuration)
	if err != nil {
		fmt.Errorf("Error parsing PomodoroWarningEndBreakDuration (%v): %v", PomodoroWarningEndBreakDuration, err)
		return nil, err
	}

	return &Pomodoro{
		session:                 session,
		guildID:                 guildID,
		textChannelID:           textChannelID,
		members:                 make(map[UserID]discordgo.User),
		status:                  PomodoroStatusStop,
		taskDuration:            taskDuration,
		breakDuration:           breakDuration,
		warningEndBreakDuration: warningEndBreakDuration,
	}, nil

}

func (p *Pomodoro) GetStatus() PomodoroStatus {
	return p.status
}

func (p *Pomodoro) Start() {
	log.Print("Pomodoro start!")

	p.taskEndTimerCh = make(chan struct{}, 1)
	p.breakEndTimerCh = make(chan struct{}, 1)
	p.stopCh = make(chan struct{}, 1)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		// infinite loop
		for {
			log.Print("Pomodoro loop!")
			select {
			case <-p.taskEndTimerCh: // end task
				p.Break()
			case <-p.breakEndTimerCh:
				p.Task()
			case <-p.stopCh:
				log.Print("Get stop channel...")
				p.timer.Stop()
				close(p.taskEndTimerCh)
				close(p.breakEndTimerCh)
				close(p.stopCh)
				log.Print("Stopped pomodoro timer!")
				return
			}
		}
	}()

	p.Task()

}

func (p *Pomodoro) Task() {
	p.status = PomodoroStatusTask

	// timer for Task

	if p.timer != nil {
		p.timer.Stop()
	}

	p.timer = time.AfterFunc(
		p.taskDuration,
		func() {
			p.taskEndTimerCh <- struct{}{}
		},
	)

	localizer := i18n.NewLocalizer(I18nBundle, language.Japanese.String())
	msg := ""

	if m, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "Start your task!"}); err == nil {
		msg += m
	} else {
		msg += "Start your task!"
	}

	msg += "\n"

	d := int(p.taskDuration.Minutes())
	if m, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "Task will end in Min minutes.",
		TemplateData: map[string]interface{}{
			"Min": d,
		},
		PluralCount: d,
	}); err == nil {
		msg += m
	} else {
		msg += "Start your task!"
	}

	msg += "\n"

	t := time.Now().Add(p.taskDuration)
	if m, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "The task will end at DateTime.",
		TemplateData: map[string]interface{}{
			"DateTime": t.Format("2006/01/02") + " " + t.Format("15:04"),
		},
	}); err == nil {
		msg += m
	} else {
		msg += "Start your task!"
	}

	log.Print(msg)
	p.messageWithAllMembersMention(msg)
	p.muteAndDeafenAllMembers()
}

func (p *Pomodoro) messageWithAllMembersMention(msg string) {
	mention := ""
	for _, user := range p.members {
		mention += "<@" + user.ID + "> "
	}
	msg = mention + "\n" + msg
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (p *Pomodoro) Break() {
	p.status = PomodoroStatusBreakTime

	// timer for break

	if p.timer != nil {
		p.timer.Stop()
	}

	localizer := i18n.NewLocalizer(I18nBundle, language.Japanese.String())

	p.timer = time.AfterFunc(
		p.breakDuration-p.warningEndBreakDuration,
		func() {
			time.AfterFunc(
				p.warningEndBreakDuration,
				func() {
					p.breakEndTimerCh <- struct{}{}
				},
			)

			// send message to all members
			var msg string
			messageID := "The break time will end soon!"
			if m, err := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: messageID,
				TemplateData: map[string]interface{}{
					"Duration": PomodoroWarningEndBreakDuration,
				},
			}); err == nil {
				msg += m
			} else {
				msg += messageID
			}
			p.messageWithAllMembersMention(msg)
		},
	)

	msg := ""
	if m, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "The break has started!"}); err == nil {
		msg += m
	} else {
		msg += "The break has started!"
	}

	msg += "\n"

	d := int(p.breakDuration.Minutes())
	if m, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "The break will end in Min minutes.",
		TemplateData: map[string]interface{}{
			"Min": d,
		},
		PluralCount: d,
	}); err == nil {
		msg += m
	} else {
		msg += "The break will end in Min minutes"
	}

	msg += "\n"

	t := time.Now().Add(p.breakDuration)

	if m, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: "The break will end at DateTime.",
		TemplateData: map[string]interface{}{
			"DateTime": t.Format("2006/01/02") + " " + t.Format("15:04"),
		},
	}); err == nil {
		msg += m
	} else {
		msg += "The break will end at DateTime."
	}

	log.Print(msg)

	p.messageWithAllMembersMention(msg)
	p.muteAndDeafenAllMembers()
	p.unMuteAndUnDeafenAllMembers()
}

func (p *Pomodoro) Stop() {
	log.Print("Trying to stop Pomodoro...")
	p.stopCh <- struct{}{}
	p.status = PomodoroStatusStop

	// wait for goroutine to finish
	p.wg.Wait()

	p.unMuteAndUnDeafenAllMembers()

	msg := "Pomodoro is over!\n"
	msg += "If you want to get out from the pomodoro VC while tasking and deaf, move to another VC from pomodoro's.\n"
	msg += "Bot can un-deafen a user only in some VC."
	log.Print(msg)
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (p *Pomodoro) AddMember(user discordgo.User) {
	p.members[user.ID] = user

	msg := "Welcome <@" + user.ID + "> !"
	switch p.status {
	case PomodoroStatusTask:
		msg += "Tasking now!"
	case PomodoroStatusBreakTime:
		msg += "Breaking now!"
	}
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}

}

func (p *Pomodoro) AddMemberWithServerMuteDeaf(user discordgo.User) {
	p.session.GuildMemberMute(p.guildID, user.ID, true)
	p.session.GuildMemberDeafen(p.guildID, user.ID, true)
	p.AddMember(user)
}

func (p *Pomodoro) RemoveMember(userID UserID) {
	p.session.GuildMemberMute(p.guildID, userID, false)
	p.session.GuildMemberDeafen(p.guildID, userID, false)
	delete(p.members, userID)
	log.Printf("Removed member: %s", userID)
}

func (p *Pomodoro) muteAndDeafenAllMembers() {
	for userID := range p.members {
		p.session.GuildMemberMute(p.guildID, userID, true)
		p.session.GuildMemberDeafen(p.guildID, userID, true)
	}
}

func (p *Pomodoro) unMuteAndUnDeafenAllMembers() {
	for userID := range p.members {
		p.session.GuildMemberMute(p.guildID, userID, false)
		p.session.GuildMemberDeafen(p.guildID, userID, false)
	}
}

// Add a new user to a pomodoro's member list
func (p *Pomodoro) AddUser(user discordgo.User) {
	switch p.GetStatus() {
	case PomodoroStatusStop:
		// Start Pomodoro
		p.AddMemberWithServerMuteDeaf(user)
		p.Start()
	case PomodoroStatusTask:
		// task中であれば入ってきた人をmute
		p.AddMemberWithServerMuteDeaf(user)
	case PomodoroStatusBreakTime:
		// 休憩中であれば入ってきた人を追加するが mute しない
		p.AddMember(user)
	}
}

var (
	pomodoroMapLock sync.Mutex
	pomodoroMap     map[GuildID]*PomodoroWithLock = make(map[GuildID]*PomodoroWithLock)
)

type PomodoroWithLock struct {
	pomo *Pomodoro
	lock sync.Mutex
}

func (pp *PomodoroWithLock) getPomodoro(session *discordgo.Session, guildID ChannelID, textChannelID ChannelID) (*Pomodoro, error) {
	pp.lock.Lock()
	log.Print("Pomodoro was locked!")
	if pp.pomo == nil {
		var err error
		if pp.pomo, err = NewPomodoro(session, guildID, textChannelID); err != nil {
			pp.lock.Unlock()
			fmt.Errorf("Error creating Pomodoro: %v", err)
			return nil, err
		}
	}
	return pp.pomo, nil
}

func getPomodoroWithLock(session *discordgo.Session, guildID GuildID, textChannelID ChannelID) (*Pomodoro, error) {
	// if empty, create a new Pomodoro
	pomodoroMapLock.Lock()
	if pomodoroMap[guildID] == nil {
		pomodoroMap[guildID] = &PomodoroWithLock{}
	}
	pomodoroMapLock.Unlock()
	return pomodoroMap[guildID].getPomodoro(session, guildID, textChannelID)
}

func unlockPomodoro(guildID GuildID) {
	pomodoroMap[guildID].lock.Unlock()
	log.Print("Pomodoro was unlocked!")
}

func releasePomodoroWithUnlock(guildID GuildID) {
	pomodoroWithLock := pomodoroMap[guildID]
	if pomodoroWithLock == nil {
		log.Printf("Pomodoro in Guild ID (%s) is not found!", guildID)
		return
	}
	lock := &pomodoroWithLock.lock
	if lock.TryLock() {
		log.Print("Pomodoro was unlocked!?")
	}
	defer unlockPomodoro(guildID)

	if pomo := pomodoroWithLock.pomo; pomo != nil {
		// Stop timer
		if pomo.status != PomodoroStatusStop {
			pomo.Stop()
		}
		// release pomodoro
		pomo = nil
		log.Printf("Pomodoro for %v was released!", guildID)
	}
}

func releaseOrUnlockPomodoro(pomodoro *Pomodoro, guildID GuildID) {
	if pomodoro == nil {
		log.Printf("Try to release or unlock pomodoro, but pomodoro is nil!")
		return
	}
	if len(pomodoro.members) == 0 {
		defer releasePomodoroWithUnlock(guildID)
	} else {
		defer unlockPomodoro(guildID)
	}
}

// 冪等性を持つ Remove User
// Lock を内部で行う
func SafeRemoveUserWithLock(guildID GuildID, userID UserID) {
	pomodoroMapLock.Lock()
	pomodoroWithLock, ok := pomodoroMap[guildID]
	if !ok { // ポモドーロが開始していなければ何もしない
		pomodoroMapLock.Unlock()
		return
	}
	pomodoroMapLock.Unlock()

	pomodoroWithLock.lock.Lock()
	pomodoro := pomodoroWithLock.pomo
	defer releaseOrUnlockPomodoro(pomodoro, guildID)
	if pomodoro == nil { // pomodoro が生成されていなければ何もしない
		return
	}

	pomodoro.RemoveMember(userID)
}

func onVoiceStateUpdate(session *discordgo.Session, updated *discordgo.VoiceStateUpdate) {
	// log.Printf("%#v", session)
	// log.Printf("onVoiceStateUpdate: %#v", updated)
	// log.Printf("%s", updated.ChannelID)

	//////////////////////////////
	// 対象のVCチャンネル以外は無視 //
	/////////////////////////////

	pomodoroVCChannelID := Info.GetChannelIDForPomodoroVC()

	if updated.ChannelID == "" && updated.BeforeUpdate.ChannelID != pomodoroVCChannelID {
		// 対象チャンネル以外からLeaveしたとき
		SafeRemoveUserWithLock(updated.GuildID, updated.UserID)
		return
	}
	if updated.BeforeUpdate == nil && updated.ChannelID != pomodoroVCChannelID {
		// どこにも入っていない状態で対象チャンネル以外にJoinしたとき
		return
	}

	// 関係ないチャンネル間の移動
	if updated.ChannelID != pomodoroVCChannelID && updated.BeforeUpdate.ChannelID != pomodoroVCChannelID {
		return
	}

	// チャンネル移動以外の変更(mute, deafen 等)は無視
	if updated.BeforeUpdate != nil && updated.BeforeUpdate.ChannelID == updated.ChannelID {
		return
	}

	user, err := session.User(updated.UserID)
	if err != nil {
		log.Print("Error getting user: ", err)
		return
	}

	//////////////
	// Pomodoro //
	//////////////

	pomodoroVCChannel, err := session.Channel(pomodoroVCChannelID)
	if err != nil {
		log.Printf("Error in get channel: %s", err)
		return
	}

	isJoin := false
	if updated.ChannelID == pomodoroVCChannel.ID {
		isJoin = true
	}
	log.Printf("%v", isJoin)

	if pomodoro, err := getPomodoroWithLock(session, pomodoroVCChannel.GuildID, Info.GetChannelIDForNotification()); err != nil {
		return
	} else {
		if isJoin {
			defer unlockPomodoro(pomodoroVCChannel.GuildID)
			pomodoro.AddUser(*user)
		} else {
			defer releaseOrUnlockPomodoro(pomodoro, pomodoroVCChannel.GuildID)
			pomodoro.RemoveMember(user.ID)
		}
	}
}
