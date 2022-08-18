package pomodoro

import (
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PomodoroStatus int

const (
	PomodoroStatusStop PomodoroStatus = iota
	PomodoroStatusTask
	PomodoroStatusBreakTime
)

const (
	PomodoroTaskMinutes  = 25
	PomodoroBreakMinutes = 5
)

// Start() でスタート
// Stop() でストップ
// Stop() 後に struct を破棄する
type Pomodoro struct {
	session       *discordgo.Session
	guildID       ChannelID
	textChannelID ChannelID
	// Joining users
	members         map[UserID]bool
	status          PomodoroStatus `default:"PomodoroStatusStop"`
	timer           *time.Timer
	taskEndTimerCh  chan struct{}
	breakEndTimerCh chan struct{}
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

func NewPomodoro(session *discordgo.Session, guildID ChannelID, textChannelID ChannelID) *Pomodoro {
	return &Pomodoro{
		session:       session,
		guildID:       guildID,
		textChannelID: textChannelID,
		members:       make(map[UserID]bool),
	}

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
		time.Duration(PomodoroTaskMinutes)*time.Minute,
		func() {
			p.taskEndTimerCh <- struct{}{}
		},
	)

	msg := "Pomodoro task was started now!"
	log.Print(msg)
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}

	p.deafenAllMembers()
}

func (p *Pomodoro) Break() {
	p.status = PomodoroStatusBreakTime

	// timer for break

	if p.timer != nil {
		p.timer.Stop()
	}

	p.timer = time.AfterFunc(
		time.Duration(PomodoroBreakMinutes)*time.Minute,
		func() {
			p.breakEndTimerCh <- struct{}{}
		},
	)

	msg := "Pomodoro break time was started now!"
	log.Print(msg)
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}

	p.unDeafenAllMembers()
}

func (p *Pomodoro) Stop() {
	log.Print("Trying to stop Pomodoro...")
	p.stopCh <- struct{}{}
	p.status = PomodoroStatusStop

	// wait for goroutine to finish
	p.wg.Wait()

	p.unDeafenAllMembers()

	msg := "Pomodoro was stopped!" + "\n" + "If you want to get out from pomodoro vc channel when tasking, move to another channel from pomodoro voice channel."
	log.Print(msg)
	if _, err := p.session.ChannelMessageSend(p.textChannelID, msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (p *Pomodoro) AddMember(userID UserID) {
	p.members[userID] = true
}

func (p *Pomodoro) AddMemberWithServerMute(userID UserID, session *discordgo.Session, guildID GuildID) {
	session.GuildMemberDeafen(guildID, userID, true)
	p.AddMember(userID)
}

func (p *Pomodoro) RemoveMember(userID UserID, session *discordgo.Session, guildID GuildID) {
	session.GuildMemberDeafen(guildID, userID, false)
	delete(p.members, userID)
	log.Printf("Removed member: %s", userID)
}

func (p *Pomodoro) deafenAllMembers() {
	for userID := range p.members {
		p.session.GuildMemberDeafen(p.guildID, userID, true)
	}
}

func (p *Pomodoro) unDeafenAllMembers() {
	for userID := range p.members {
		p.session.GuildMemberDeafen(p.guildID, userID, false)
	}
}

var (
	pomodoroMapLock sync.Mutex
	pomodoroMap     map[ChannelID]*PomodoroWithLock = make(map[ChannelID]*PomodoroWithLock)
)

type PomodoroWithLock struct {
	pomo *Pomodoro
	lock sync.Mutex
}

func (pp *PomodoroWithLock) getPomodoro(session *discordgo.Session, guildID ChannelID, textChannelID ChannelID) *Pomodoro {
	pp.lock.Lock()
	log.Print("Pomodoro was locked!")
	if pp.pomo == nil {
		pp.pomo = NewPomodoro(session, guildID, textChannelID)
	}
	return pp.pomo
}

func getPomodoroWithLock(session *discordgo.Session, guildID ChannelID, vcChannelID ChannelID, textChannelID ChannelID) *Pomodoro {
	// if empty, create a new Pomodoro
	pomodoroMapLock.Lock()
	if pomodoroMap[vcChannelID] == nil {
		pomodoroMap[vcChannelID] = &PomodoroWithLock{}
	}
	pomodoroMapLock.Unlock()
	return pomodoroMap[vcChannelID].getPomodoro(session, guildID, textChannelID)
}

func unlockPomodoro(channelID ChannelID) {
	pomodoroMap[channelID].lock.Unlock()
	log.Print("Pomodoro was unlocked!")
}

func releasePomodoroWithUnlock(channelID ChannelID) {
	pomodoroWithLock := pomodoroMap[channelID]
	if pomodoroWithLock == nil {
		log.Printf("Pomodoro with channelID: %s is not found!", channelID)
		return
	}
	lock := &pomodoroWithLock.lock
	if lock.TryLock() {
		log.Print("Pomodoro was unlocked!?")
	}
	defer unlockPomodoro(channelID)

	if pomo := pomodoroWithLock.pomo; pomo != nil {
		// Stop timer
		if pomo.status != PomodoroStatusStop {
			pomo.Stop()
		}
		// release pomodoro
		pomo = nil
		log.Printf("Pomodoro for %v was released!", channelID)
	}
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

	pomodoro := getPomodoroWithLock(session, pomodoroVCChannel.GuildID, pomodoroVCChannelID, Info.GetChannelIDForNotification())
	if isJoin {
		defer unlockPomodoro(pomodoroVCChannel.ID)
	} else {
		pomodoro.RemoveMember(user.ID, session, pomodoroVCChannel.GuildID)
		if len(pomodoro.members) == 0 {
			defer releasePomodoroWithUnlock(pomodoroVCChannel.ID)
		} else {
			defer unlockPomodoro(pomodoroVCChannel.ID)
		}
		return
	}

	switch pomodoro.GetStatus() {
	case PomodoroStatusStop:
		// Start Pomodoro
		pomodoro.Start()
		pomodoro.AddMemberWithServerMute(user.ID, session, pomodoroVCChannel.GuildID)
	case PomodoroStatusTask:
		// task中であれば入ってきた人をmute
		pomodoro.AddMemberWithServerMute(user.ID, session, pomodoroVCChannel.GuildID)
	case PomodoroStatusBreakTime:
		// 休憩中であれば入ってきた人を追加するが mute しない
		pomodoro.AddMember(user.ID)
	}

}