# NovaUserbot

![NovaUserbot Logo](https://files.indrajeeth.in/nova.jpg)

NovaUserbot is a powerful and flexible userbot written in Go. It leverages the Telegram API to automate tasks and enhance your Telegram experience.

## Features

- 🤖 Easy to set up and use
- 🚀 High performance with Go
- 🔒 Secure and reliable
- 📚 Built using the Gogram library
- 🗄️ Uses Redis for fast database management
- 🌐 Multi-language support (English, Hindi)
- 🔧 Configurable via database variables
- 👥 Multi-user support with sudo system

## Requirements

- Go 1.23 or higher
- Redis database
- Telegram API credentials

## Quick Start

### 1. Get Telegram API Credentials

1. Go to [https://my.telegram.org](https://my.telegram.org)
2. Log in with your phone number
3. Go to "API development tools"
4. Create a new application
5. Copy your `API_ID` and `API_HASH`

### 2. Generate String Session

You have two options to generate a string session:

#### Option A: Online Generator (Recommended)
Visit [https://sess.gogram.fun/](https://sess.gogram.fun/) and follow the instructions.

#### Option B: Local Generator
```sh
cd scripts
go mod init session_gen
go mod tidy
go run session_gen.go
```

### 3. Get a Bot Token

1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Send `/newbot` and follow the instructions
3. Copy your bot token

### 4. Set Up Redis

You can use a free Redis instance from:
- [Redis Cloud](https://redis.com/try-free/)
- [Upstash](https://upstash.com/)
- Or run locally: `docker run -d -p 6379:6379 redis`

---

## Deployment Options

### 🐳 Docker Deployment (Recommended)

#### Using Docker Compose

1. Clone the repository:
```sh
git clone https://github.com/IndrajeethY/Nova.git
cd NovaUserbot
```

2. Create a `.env` file:
```sh
cp .env.example .env
# Edit .env with your credentials
```

3. Start with Docker Compose:
```sh
docker-compose up -d
```

4. View logs:
```sh
docker-compose logs -f
```

5. Stop the bot:
```sh
docker-compose down
```

#### Using Docker directly

1. Build the image:
```sh
docker build -t novauserbot .
```

2. Run the container:
```sh
docker run -d \
  --name novauserbot \
  -e API_ID=your_api_id \
  -e API_HASH=your_api_hash \
  -e TOKEN=your_bot_token \
  -e STRING_SESSION=your_string_session \
  -e DB_URL=redis://username:password@host:port \
  --restart unless-stopped \
  novauserbot
```

---

### 💻 Local Deployment

#### Prerequisites
- Go 1.23 or higher installed
- Git installed
- Redis running locally or accessible remotely

#### Steps

1. Clone the repository:
```sh
git clone https://github.com/IndrajeethY/Nova.git
cd NovaUserbot
```

2. Install dependencies:
```sh
go mod tidy
```

3. Set up environment variables:

**Linux/macOS:**
```sh
export API_ID=your_api_id
export API_HASH=your_api_hash
export TOKEN=your_bot_token
export STRING_SESSION=your_string_session
export DB_URL=redis://localhost:6379
```

**Windows (PowerShell):**
```powershell
$env:API_ID="your_api_id"
$env:API_HASH="your_api_hash"
$env:TOKEN="your_bot_token"
$env:STRING_SESSION="your_string_session"
$env:DB_URL="redis://localhost:6379"
```

**Or create a `.env` file:**
```sh
API_ID=your_api_id
API_HASH=your_api_hash
TOKEN=your_bot_token
STRING_SESSION=your_string_session
DB_URL=redis://localhost:6379
```

4. Run the bot:
```sh
go run .
```

5. Or build and run:
```sh
go build -o novauserbot
./novauserbot
```

---

### ☁️ Cloud Deployment

#### Heroku

1. Create a new Heroku app
2. Add Redis addon: `heroku addons:create heroku-redis:hobby-dev`
3. Set config vars:
```sh
heroku config:set API_ID=your_api_id
heroku config:set API_HASH=your_api_hash
heroku config:set TOKEN=your_bot_token
heroku config:set STRING_SESSION=your_string_session
```
4. Deploy:
```sh
git push heroku main
```

#### Railway

1. Fork this repository
2. Create a new project on [Railway](https://railway.app/)
3. Connect your GitHub repo
4. Add environment variables
5. Deploy!

---

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `API_ID` | ✅ | Telegram API ID | `123456` |
| `API_HASH` | ✅ | Telegram API Hash | `abcdef1234567890` |
| `TOKEN` | ✅ | Bot token from BotFather | `123456:ABC-DEF` |
| `STRING_SESSION` | ✅ | String session for userbot | `1BvXWG...` |
| `DB_URL` | ✅ | Redis connection URL | `redis://user:pass@host:port` |

---

## Database Variables

You can configure these variables using `.setvar` command:

| Variable | Description | Default |
|----------|-------------|---------|
| `CMD_HANDLER` | Command prefix | `.` |
| `ALIVE_IMAGE` | Custom alive image URL | - |
| `LOG_CHAT` | Chat ID for logging | - |
| `BOT_LANGUAGE` | Bot language (en/hi) | `en` |
| `GIT_TOKEN` | GitHub token for updates | - |
| `UPSTREAM_REPO` | Upstream repo URL | Default repo |
| `UPSTREAM_BRANCH` | Upstream branch | `main` |
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `PM_AI_PROMT` | Custom PM assistant prompt | - |

---

## Google Drive Setup

To use the Google Drive module, follow these steps:

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click on "Select a project" → "New Project"
3. Enter a project name and click "Create"

### 2. Enable Google Drive API

1. In your project dashboard, go to "APIs & Services" → "Library"
2. Search for "Google Drive API"
3. Click on it and press "Enable"

### 3. Create OAuth 2.0 Credentials

1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth client ID"
3. If prompted, configure the OAuth consent screen:
   - Choose "External" user type
   - Fill in the required fields (App name, User support email, Developer email)
   - Add scope: `https://www.googleapis.com/auth/drive`
   - Add your email as a test user
4. Create OAuth client ID:
   - Application type: "Desktop app"
   - Name: "NovaUserbot" (or any name)
   - Click "Create"
5. Copy your **Client ID** and **Client Secret**

### 4. Configure in Userbot

1. Send this command to set up credentials:
   ```
   .gsetup <client_id> <client_secret>
   ```

2. You'll receive an authorization URL. Open it in your browser.

3. Sign in with your Google account and authorize the app.

4. Copy the authorization code and use:
   ```
   .gauth <authorization_code>
   ```

5. Done! You can now use Google Drive commands.

### Google Drive Commands

| Command | Description |
|---------|-------------|
| `.gsetup <client_id> <client_secret>` | Setup Google Drive credentials |
| `.gauth <code>` | Complete authorization with auth code |
| `.gupload` | Upload replied file to Google Drive |
| `.glist` | List files in your Drive |
| `.gsearch <query>` | Search files in Drive |
| `.gdown <file_id/link>` | Download file from Drive |
| `.gdelete <file_id/link>` | Delete file from Drive |

---

## Commands

### Core
| Command | Description |
|---------|-------------|
| `.help` | Show help menu or search modules/commands |
| `.ping` | Ping the userbot |
| `.dcping` | Ping all data centers |
| `.alive` | Check if bot is running |

### Admin
| Command | Description |
|---------|-------------|
| `.ban` | Ban a user from the chat |
| `.unban` | Unban a user from the chat |
| `.kick` | Kick a user from the chat |
| `.mute` | Mute a user in the chat |
| `.unmute` | Unmute a user in the chat |
| `.dmute` | Delete message and mute a user |
| `.dkick` | Delete message and kick a user |
| `.dban` | Delete message and ban a user |
| `.promote` | Promote a user to admin |
| `.fullpromote` | Fully promote a user with all rights |
| `.demote` | Demote a user from admin |
| `.pin` | Pin a message in the chat |
| `.unpin` | Unpin a message |
| `.zombies` | Find and clean deleted accounts |

### BanGuard
| Command | Description |
|---------|-------------|
| `.gconfig <duration> <limit>` | Set BanGuard limits (e.g., `.gconfig 10s 5`) |
| `.gtoggle on/off` | Toggle BanGuard on/off |
| `.gstatus` | Check BanGuard status |

### Gban
| Command | Description |
|---------|-------------|
| `.gban` | Globally ban a user |
| `.ungban` | Globally unban a user |
| `.gbanned` | List globally banned users |
| `.antispam` | Toggle antispam |

### Sudoers
| Command | Description |
|---------|-------------|
| `.addsudo` | Add user as sudo |
| `.delsudo` | Remove user from sudo |
| `.listsudo` | List all sudos |

### PM Permit
| Command | Description |
|---------|-------------|
| `.ap` | Approve a user |
| `.dap` | Disapprove a user |
| `.approved` | List approved users |
| `.setprompt` | Set PM assistant prompt |

### AFK
| Command | Description |
|---------|-------------|
| `.afk [reason]` | Set AFK status (supports media reply) |
| `.brb [reason]` | Alias for `.afk` |

### Database
| Command | Description |
|---------|-------------|
| `.setvar <key> <value>` | Set a database variable |
| `.getvar <key>` | Get a database variable |
| `.delvar <key>` | Delete a database variable |
| `.vars` | List all database variables |
| `.delallvars` | Delete all variables (requires confirm) |

### Files
| Command | Description |
|---------|-------------|
| `.file <file_id>` | Send a file by its FileID |
| `.fid` | Get FileID of replied media |
| `.ul` | Upload a file (`-s` for spoiler, `--doc` for document) |
| `.dl` | Download replied file or from link |
| `.cancel` | Cancel an active download |
| `.finfo` | Get file information |

### File Sharing
| Command | Description |
|---------|-------------|
| `.share [service]` | Share to hosting (catbox/gofile/fileio/0x0) |
| `.catbox` | Upload to Catbox.moe |
| `.gofile` | Upload to GoFile.io |
| `.fileio` | Upload to File.io (one-time download) |
| `.genlink` | Generate link for media |

### Google Drive
| Command | Description |
|---------|-------------|
| `.gsetup <client_id> <client_secret>` | Setup Google Drive credentials |
| `.gauth <code>` | Complete authorization with auth code |
| `.gupload` | Upload replied file to Google Drive |
| `.glist` | List files in Google Drive |
| `.gsearch <query>` | Search files in Google Drive |
| `.gdown <file_id/link>` | Download file from Google Drive |
| `.gdelete <file_id/link>` | Delete file from Google Drive |

### Image Tools
| Command | Description |
|---------|-------------|
| `.grey` | Convert image to grayscale |
| `.blur` | Apply blur effect to image |
| `.negative` | Create negative of image |
| `.mirror` | Mirror image horizontally |
| `.flip` | Flip image vertically |
| `.rotate [angle]` | Rotate image (default: 90) |
| `.sketch` | Convert image to sketch |
| `.border <color> ; <width>` | Add border to image |
| `.pixelate [scale]` | Pixelate image (1-100) |
| `.sepia` | Apply sepia tone effect |
| `.emboss` | Apply emboss effect |
| `.sharpen` | Sharpen image |
| `.resize <WxH or %>` | Resize image |
| `.csample <color>` | Create color sample |

### Media Tools
| Command | Description |
|---------|-------------|
| `.mediainfo` | Get media file information |
| `.vrotate <angle>` | Rotate video/image (90/180/270) |
| `.vcompress [crf]` | Compress video (CRF 0-51) |
| `.vtogif` | Convert video to GIF |
| `.giftov` | Convert GIF to video |
| `.vtrim <start> <end>` | Trim video (HH:MM:SS format) |

### Paste
| Command | Description |
|---------|-------------|
| `.paste` | Paste to Nekobin (`-p` for other services) |
| `.read` | Read file content (`-n` for line limit) |

### Profile
| Command | Description |
|---------|-------------|
| `.setname <name>` | Change profile name (use `//` for last name) |
| `.setbio <bio>` | Change profile bio |
| `.setpic` | Change profile picture (reply to media) |
| `.delpfp [n/all]` | Delete profile picture(s) |
| `.poto [user] [count]` | Get profile picture(s) |

### Stories
| Command | Description |
|---------|-------------|
| `.setstory [all/contacts]` | Set replied media as story |
| `.storydl <user/link>` | Download user stories |
| `.archdl [-n <index>]` | Download archived story |

### Search
| Command | Description |
|---------|-------------|
| `.github <username>` | Get GitHub user profile info |
| `.gh <username>` | Alias for `.github` |
| `.google <query>` | Search on Google/DuckDuckGo |
| `.img <query>` | Search for images |
| `.imdb <query>` | Search movies/series on IMDB |
| `.unsplash <query> ; [count]` | Search and download Unsplash images |

### User Info
| Command | Description |
|---------|-------------|
| `.info` | Fetch user info |
| `.id` | Fetch ID info |
| `.stats` | Fetch user statistics |

### Reminders
| Command | Description |
|---------|-------------|
| `.remind <time> <message>` | Set a reminder (e.g., `.remind 1h30m Buy groceries`) |
| `.reminders` | List your reminders |
| `.delreminder <index>` | Delete a reminder by index |
| `.clearreminders` | Clear all your reminders |

### Logging
| Command | Description |
|---------|-------------|
| `.setlog` | Set log channel |
| `.getlog` | Get log channel |
| `.dellog` | Delete log channel |
| `.logging` | Toggle logging on/off |
| `.taglogger` | Set tag logger chat |
| `.gettaglogger` | Get tag logger chat |
| `.deltaglogger` | Delete tag logger |

### ChatBot
| Command | Description |
|---------|-------------|
| `.ai <query>` | Query Gemini AI |

### SpeedTest
| Command | Description |
|---------|-------------|
| `.speedtest` | Run a network speedtest |
| `.st` | Alias for `.speedtest` |

### Language
| Command | Description |
|---------|-------------|
| `.setlang <code>` | Set bot language (en/hi) |
| `.lang` | Show current language |

### Updater
| Command | Description |
|---------|-------------|
| `.checkupdate` | Check for updates |
| `.update` | Update to latest version |

### Developer
| Command | Description |
|---------|-------------|
| `.sh <command>` | Execute shell command |
| `.eval <code>` | Evaluate Go code |
| `.json` | Get message JSON |

---

## Multi-Language Support

NovaUserbot supports multiple languages. Currently available:
- English (en)
- Hindi (hi)

Set your preferred language:
```
.setvar BOT_LANGUAGE hi
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## Credits

This project might contain combination of modules from various userbots. Full credit to the original authors of those modules.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
