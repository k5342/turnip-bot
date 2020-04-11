package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"bufio"
	"io"
	"errors"

	"github.com/bwmarrin/discordgo"
)

type UserRecord struct {
	Prices      [12]uint16
	BuyingPrice uint16
}

// cache tables
var usernames map[string]discordgo.User
var invitedChannelIDs map[string]bool
var userRecords map[string]*UserRecord
var channelTZs map[string]string
var channelTZcache map[string]*time.Location
var priceSlots []string

func main() {
	dg, err := discordgo.New("Bot " + os.Getenv("TURNIPBOT_TOKEN"))
	if err != nil {
		fmt.Println("Error creating Discord bot: ", err)
		return
	}

	usernames = make(map[string]discordgo.User)
	invitedChannelIDs = make(map[string]bool)
	userRecords = make(map[string]*UserRecord)
	channelTZs = make(map[string]string)
	channelTZcache = make(map[string]*time.Location)
	priceSlots = []string{
		"Mon/AM", "Mon/PM", "Tue/AM", "Tue/PM", "Wed/AM", "Wed/PM",
		"Thu/AM", "Thu/PM", "Fri/AM", "Fri/PM", "Sat/AM", "Sat/PM",
		"Sun",
	}
	dg.AddHandler(messageHandler)
	dg.AddHandler(userPresenceUpdateHandler)

	dg.Open()
	if err != nil {
		fmt.Println("Error opening WebSocket connection: ", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func sendReply(s *discordgo.Session, m *discordgo.MessageCreate, str string) {
	sendMessage := fmt.Sprintf("<@!%s> ", m.Author.ID)
	sendMessage += str
	s.ChannelMessageSend(m.ChannelID, sendMessage)
}

func isContain(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func userPresenceUpdateHandler(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	if p.User.Username != "" {
		// update cache
		usernames[p.User.ID] = *p.User
	}
}

func getLabelFromPriceIndex(price_index int) string {
	return priceSlots[price_index]
}

func checkInvited(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	return (invitedChannelIDs[m.ChannelID] == true)
}

func doInvite(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	invitedChannelIDs[m.ChannelID] = true
	sendReply(s, m, "Successfully invited!!")
}

func showList(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	user_id := m.Author.ID
	_ = restoreUserRecord(m.Author.ID, m.ChannelID)
	if r, ok := userRecords[user_id]; ok {
		result := "This week: \n"
		if r.BuyingPrice > 0 {
			result += fmt.Sprintf("BuyingPrice: %d", r.BuyingPrice)
		}
		for i := range r.Prices {
			if r.Prices[i] > 0 {
				result += fmt.Sprintf("%s: %d\n", priceSlots[i], r.Prices[i])
			}
		}
		sendReply(s, m, result)
		return
	} else {
		sendReply(s, m, "no records found")
		return
	}
}

func saveRecord(user_id string, record_index int, price int, ts time.Time) error {
	if _, ok := userRecords[user_id]; !ok {
		userRecords[user_id] = new(UserRecord)
	}
	
	// write log
	f, err := os.OpenFile("./record-log.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	ts_week := getLastSunday(ts).Format("2006-01-02")
	ts_submitted := ts.String()
	_, err = f.WriteString(strings.Join([]string{user_id, strconv.Itoa(record_index), strconv.Itoa(price), ts_week, ts_submitted}, "\t") + "\n")
	if err != nil {
		return err
	}
	f.Close()
	
	if record_index < 12 {
		userRecords[user_id].Prices[record_index] = uint16(price)
	} else {
		userRecords[user_id].BuyingPrice = uint16(price)
	}
	return nil
}

func restoreUserRecord(user_id string, channel_id string) (error) {
	ts, err := getChannelLocalTime(time.Now(), channel_id)
	if err != nil {
		return err
	}
	ts_week_current := getLastSunday(ts).Format("2006-01-02")
	f, err := os.Open("./record-log.csv")
	if err != nil {
		return err
	}
	
	reader := bufio.NewReaderSize(f, 2048)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line = strings.TrimRight(line, "\n")
		c := strings.Split(line, "\t")
		if c[0] != user_id {
			continue
		}
		ts_week := c[3]
		if err != nil {
			return err
		}
		fmt.Println(ts_week_current)
		fmt.Println(ts_week)
		fmt.Println(line)
		if ts_week_current == ts_week {
			record_index, err := strconv.Atoi(c[1])
			if err != nil {
				return err
			}
			if record_index >= 0 && record_index <= 12 {
				r := new(UserRecord)
				price, err := strconv.Atoi(c[2])
				if err != nil {
					return err
				}
				if !isValidPrice(price) {
					return errors.New("Price Validation Error")
				}
				if record_index < 12 {
					r.Prices[record_index] = uint16(price)
				} else {
					r.BuyingPrice = uint16(price)
				}
				userRecords[user_id] = r
			} else {
				continue
			}
		}
	}
	return nil
}

func isValidPrice(p int) bool {
	if p >= 65536 || p < 0 {
		return false
	}
	return true
}

func doAppendData(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	if checkInvited(s, m) != true {
		return
	}
	if len(args) == 0 {
		sendReply(s, m, fmt.Sprintf("Usage: `%s add [price] <wday> <am/pm>`", prefix))
		return
	} else {
		timeobj, errl := parseLocalTime(m)
		if errl != nil {
			sendReply(s, m, fmt.Sprintf("add: week detection failed"))
			return
		}
		price, errp := strconv.Atoi(args[0])
		if errp != nil {
			sendReply(s, m, fmt.Sprintf("wrong price format"))
			return
		}
		var record_index int
		if len(args) == 1 {
			// auto complete wday and time
			record_index, err := getRecordIndex(m)
			if err != nil {
				sendReply(s, m, fmt.Sprintf("add: cannot detect current time"))
				return
			}
			err = saveRecord(string(m.Author.ID), record_index, price, timeobj)
			if err != nil {
				sendReply(s, m, fmt.Sprintf("add: failed to save your record"))
			}
			sendReply(s, m, fmt.Sprintf("record_index: %d, price: %d, %s", record_index, price, timeobj.String()))
			return
		} else if len(args) <= 3 {
			// parse second, third arguments
			if strings.ToLower(args[1]) == "sun" {
				err := saveRecord(string(m.Author.ID), 12, price, timeobj)
				if err != nil {
					sendReply(s, m, fmt.Sprintf("add: failed to save your record"))
				}
				sendReply(s, m, fmt.Sprintf("buying_price: %d", price))
			} else {
				if len(args) == 3 {
					table := []string{"mon", "tue", "wed", "thu", "fri", "sat", "XXX"}
					for i, v := range table {
						if v == strings.ToLower(args[1]) {
							record_index = 2 * i
							break
						}
						if len(table) - 1 == i {
							sendReply(s, m, "wday is invalid: please specify from mon, tue, wed, thu, fri, sat")
							return
						}
					}
					if strings.ToLower(args[2]) == "am" {
						record_index += 0
					} else if strings.ToLower(args[2]) == "pm" {
						record_index += 1
					} else {
						sendReply(s, m, "am/pm is invalid: please specify from am, pm")
						return
					}
					
					// save
					err := saveRecord(string(m.Author.ID), record_index, price, timeobj)
					if err != nil {
						sendReply(s, m, fmt.Sprintf("add: failed to save your record"))
					}
					sendReply(s, m, fmt.Sprintf("Registered: %s: %d Bell", getLabelFromPriceIndex(record_index), price))
					return
				} else {
					sendReply(s, m, fmt.Sprintf("please specify both wday and am/pm"))
					return
				}
			}
		} else {
			// invalid arguments
			sendReply(s, m, fmt.Sprintf("add: invalid arguments"))
			return
		}
	}
}

func getRecordIndex(m *discordgo.MessageCreate) (int, error) {
	timeobj, err := parseLocalTime(m)
	if err != nil {
		return -1, err
	}
	if timeobj.Hour() < 5 {
		return 2 * ((int(timeobj.Weekday()) + 6) % 7 - 1 /* yesterday */) + 1 /* pm */, nil
	} else if timeobj.Hour() < 12 {
		return 2 * ((int(timeobj.Weekday()) + 6) % 7 + 0 /* today */) + 0 /* am */, nil
	} else {
		return 2 * ((int(timeobj.Weekday()) + 6) % 7 + 0 /* today */) + 1 /* pm */, nil
	}
}

func parseLocalTime(m *discordgo.MessageCreate) (time.Time, error) {
	timeobj, err := m.Timestamp.Parse()
	if err != nil {
		return time.Now(), err
	}
	return getChannelLocalTime(timeobj, m.ChannelID)
}

func getChannelLocalTime(t time.Time, channel_id string) (time.Time, error) {
	if channelTZcache[channel_id] == nil {
		return t.In(time.Local), nil
	} else {
		return t.In(channelTZcache[channel_id]), nil
	}
}

func getLastSunday(ts time.Time) time.Time {
	switch ts.Weekday() {
	case time.Sunday:
		return ts
	case time.Monday:
		return ts.AddDate(0, 0, -1)
	case time.Tuesday:
		return ts.AddDate(0, 0, -2)
	case time.Wednesday:
		return ts.AddDate(0, 0, -3)
	case time.Thursday:
		return ts.AddDate(0, 0, -4)
	case time.Friday:
		return ts.AddDate(0, 0, -5)
	case time.Saturday:
		return ts.AddDate(0, 0, -6)
	}
	return time.Now()
}

func setTimezone(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	if len(args) == 0 {
		sendReply(s, m, fmt.Sprintf("Usage: `%s [IANA_TZ]`", prefix))
		return
	}
	location, err := time.LoadLocation(args[0])
	if err != nil {
		sendReply(s, m, fmt.Sprintf("`%s` is not available", args[0]))
		return
	}
	channelTZs[m.ChannelID] = args[0]
	channelTZcache[m.ChannelID] = location
	localtime, err := parseLocalTime(m)
	if err == nil {
		sendReply(s, m, fmt.Sprintf("timezone has successfully updated: %s", localtime.String()))
	} else {
		sendReply(s, m, fmt.Sprintf("unexpected error while processing timezone"))
	}
}

func doAppendDataShort(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	if checkInvited(s, m) != true {
		return
	}
	if len(args) == 0 {
		return
	}
	// auto complete wday and time
	timeobj, err1 := parseLocalTime(m)
	record_index, err2 := getRecordIndex(m)
	if err1 != nil && err2 != nil {
		sendReply(s, m, fmt.Sprintf("add: cannot detect current time"))
		return
	}
	prices := strings.Split(args[0], "/")
	if len(prices) == 2 {
		if record_index >= 12 {
			sendReply(s, m, "slashed short format is disabled on Sunday")
			return
		}
		price_am, erra := strconv.Atoi(prices[0])
		price_pm, errp := strconv.Atoi(prices[1])
		if erra != nil || errp != nil {
			return
		}
		if !isValidPrice(price_am) || !isValidPrice(price_pm) {
			return
		}
		errwa := saveRecord(string(m.Author.ID), record_index - 1, price_am, timeobj)
		errwp := saveRecord(string(m.Author.ID), record_index - 0, price_pm, timeobj)
		if errwa != nil || errwp != nil {
			sendReply(s, m, fmt.Sprintf("add: failed to save your record"))
		}
		sendReply(s, m, fmt.Sprintf("record_index: %d, price: %d/%d, %s", record_index, price_am, price_pm, timeobj.String()))
	} else {
		price, err := strconv.Atoi(args[0])
		if err != nil {
			return
		}
		if !isValidPrice(price) {
			return
		}
		err = saveRecord(string(m.Author.ID), record_index, price, timeobj)
		if err != nil {
			sendReply(s, m, fmt.Sprintf("add: failed to save your record"))
		}
		sendReply(s, m, fmt.Sprintf("Registered: %s: %d Bell", getLabelFromPriceIndex(record_index), price))
	}
}

func showUsage(s *discordgo.Session, m *discordgo.MessageCreate, prefix string, args []string) {
	localtime, errlt := parseLocalTime(m)
	price_index, errct := getRecordIndex(m)
	if errlt == nil && errct == nil {
		replyStr := "\n" + 
			fmt.Sprintf("Current Time: **%s (%s)**\n", localtime.String(), getLabelFromPriceIndex(price_index)) +
			fmt.Sprintf("`%s help` ... show command help\n", prefix) +
			fmt.Sprintf("`%s invite` ... Invite me to this channel\n", prefix) +
			fmt.Sprintf("`%s timezone [TZ]` ... Set timezone for this channel\n", prefix)
		if checkInvited(s, m) == true {
			replyStr +=
				fmt.Sprintf("`%s add [price] <wday> <am/pm>` ... add the price data\n", prefix) +
				fmt.Sprintf("`%s list` ... show your prices on this week\n", prefix) + 
				fmt.Sprintf("`%s graph` ... visualize your prices on this week\n", prefix) + 
				fmt.Sprintf("`%s reminder [set|check|list|rm]` ... configure reminder\n", prefix)
		} else {
			replyStr +=
				fmt.Sprintf("*.. some features are available after invitation .. type `%s invite` ..*\n", prefix)
		}
		sendReply(s, m, replyStr);
	} else {
		sendReply(s, m, "service unavailable: cannot fetch current time")
	}
}
func parseCommand(args []string) (func(*discordgo.Session, *discordgo.MessageCreate, string, []string), []string) {
	if args[0] == "invite" {
		return doInvite, args[1:]
	}
	if args[0] == "help" {
		return showUsage, args[1:]
	}
	if args[0] == "list" {
		return showList, args[1:]
	}
	if args[0] == "add" {
		return doAppendData, args[1:]
	}
	if args[0] == "timezone" {
		return setTimezone, args[1:]
	}
	return showUsage, args[1:]
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// parse command
	if strings.HasPrefix(m.Content, "!!turnip") ||
		strings.HasPrefix(m.Content, "!!kabu") {
		args := strings.Split(m.Content, " ")
		if len(args) >= 2 {
			func_to_run, func_args := parseCommand(args[1:])
			func_to_run(s, m, args[0], func_args)
			return
		} else {
			showUsage(s, m, args[0], args)
			return
		}
	} else if _, err := strconv.Atoi(m.Content); err == nil {
		doAppendDataShort(s, m, "", []string{m.Content} )
		return
	} else if strings.Contains(m.Content, "/") {
		doAppendDataShort(s, m, "", []string{m.Content} )
		return
	} else {
		// ignore
		return
	}
}
