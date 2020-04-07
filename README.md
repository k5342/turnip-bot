# turnip-bot
An awesome discord bot to track the price of *turnip* on your island/village.
turnip-bot memorize the prices by users your Discord server and visualize it.

## Features
*turnip-bot* adds these features into any channel in your Discord server.
- [WIP] Track your price data
- [WIP] Price pattern detection
- [TODO] Set reminder(s)
- [TODO] Visualize recent data
- [TODO] Show your price history

## Usage
1. Invite turnip-bot to your preferred text channel  -- `!!turnip invite`
1. Type current price of turnip  -- ex. `!!turnip add 103`
1. Customize the bot if necessary. Check the command reference below.

### Invite
Invite turnip-bot using following command:
```
!!turnip invite
```

### Input data
```
!!turnip add [price] <wday> <am/pm>
```

- **price**: The price of your turnip (Integer, Required)
- **wday**: `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`
  - When you use `sun`, the bot tracks your buying price. The next argument `am/pm` is ignored.
- **am/pm**: `am`, `pm`


You can also use *short format* shows below;
```
DDD
AAA/PPP
```

- **DDD**: An integer. The current price. The turnip-bot automatically recognize AM or PM by current time.  (ex: When your price is 90 bell, you type `90`. )
- **AAA/PPP**: Two numbers paired in slash. Indicates the prices today (both AM/PM). The turnip-bot automatically recognize the date. (ex: When the price 90 and 110, respectively, you type `90/110`).

## Command Reference

| Command | Description |
|:--------|:------------|
|`!!turnip`| Dump command usage |
|`!!turnip invite`| Invite turnip-bot to the channel you typed this command |
|`!!turnip list`| Dump your prices on this week |
|`!!turnip list <screen_name>`| Dump prices of *screen_name* on this week |
|`!!turnip graph`| Visualize your prices on this week |
|`!!turnip graph <screen_name>`| Visualize your prices on this week |
|`!!turnip reminder set`| Set a reminder on both 8:00 am / 12:00 pm. on your timezone |
|`!!turnip reminder check`| Print the reminder config on this channel |
|`!!turnip reminder list`| Print the reminder config on this server |
|`!!turnip reminder set <0-24>`| Set a reminder at you specified hour |
|`!!turnip reminder rm all`| Remove all reminder from your channel |
|`!!turnip reminder rm <0-24>`| Remove the specified reminder from your channel |
|`!!turnip lang <lang>`| Set language |

## Installation and Setup

### On Dedicated Host

#### Clone & Install
To run manually, clone this repositoy and install dependency:
```
go get github.com/bwmarrin/discordgo
go build main.go
```

#### Register an Application on Discord Developer Portal
And create an *Application* (bot feature and some permissions enabled) and issue *Token* from [Discord Developer Portal](https://discordapp.com/developers/applications/)

This bot requires following *Bot Permissions*:
```
- Send Messages
```

#### Run
Set *Token* as `TURNIPBOT_TOKEN` environment variable.
```
export TURNIPBOT_TOKEN=`XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
```

Or you are on Windows:
```
set TURNIPBOT_TOKEN=`XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
```

Finally run the bot.
```
./main
.\main.exe
```

#### Invite to your server
And to invite your server, generate invitation URL manually using *Client ID* on *Discord Developer Portal*.
You can create Invitation URL manually following this template. More details are available at [discord documentation](https://discordapp.com/developers/docs/topics/oauth2#bots).
```
https://discordapp.com/oauth2/authorize?client_id=<client_id>&scope=bot&permissions=2048
```

### From This Invitation Link (Beta)
**Disclaimer:** We may stop this service at any time on no notice and no warranty. And since this bot can view some messages on your server, please set appropriate permissions to protect from your sensitive messages and unexpected malfunction!!  
**Privacy Policy: The data you sent through the bot is stored our database with your Discord client id. Please refrain from using our hosting if you want to avoid this. We may see or use the database for debug or feature release.**

```
https://discordapp.com/oauth2/authorize?client_id=696740526644265042&scope=bot&permissions=2048
```

## LICENSE
The MIT License

Copyright 2020 k5342

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.