# ZetaComms

ZetaComms is a notification service for ZetaChain governance proposals. It monitors proposals on different networks and sends notifications to various communication channels including Slack, Discord, and Telegram.

## Features

- Monitor ZetaChain governance proposals
- Send notifications to multiple channels:
  - Slack (via webhooks)
  - Discord (via bot)
  - Telegram (via bot)
- Configurable audiences and channels
- Rich message formatting with status indicators

## Installation

```bash
# Clone the repository
git clone https://github.com/hazim1093/zeta-comms.git
cd zeta-comms

# Install dependencies
go mod download
```

## Configuration

Configuration is managed through a YAML file located at `configs/config.yaml`. You can also provide additional configuration files using the `--config` flag.

### Basic Configuration

```yaml
logging:
  level: info
  format: console # console/text or json

discord:
  bot_token: "${DISCORD_BOT_TOKEN}" # Use environment variable for security

telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN}" # Use environment variable for security

networks:
  mainnet:
    api_url: https://mainneturl.com
    poll_interval: 1m
    audiences:
      - mainnet_operators
      - testnet_operators
  testnet:
    api_url: https://zetachain-athens.blockpi.network/lcd/v1/public
    poll_interval: 5s
    audiences:
      - testnet_operators

audience_config:
  mainnet_operators:
    channels:
      discord:
        - "123456789012345678" # Discord channel ID
      slack:
        - https://slack.com/mainnet-webhook
      telegram:
        - "-1001234567890" # Telegram chat ID for mainnet group
  testnet_operators:
    channels:
      discord:
        - "876543210987654321" # Discord channel ID
      slack:
        - https://slack.com/testnet
      telegram:
        - "-1005678901234" # Telegram chat ID for testnet

storage:
  filename: file-db.yaml
```

## Setting Up Discord Bot

To use Discord notifications, you need to create a Discord bot and add it to your server. Follow these steps:

### 1. Create a Discord Application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and give it a name (e.g., "ZetaChain Notifications")
3. Navigate to the "Bot" tab and click "Add Bot"
4. Under the "TOKEN" section, click "Copy" to get your bot token
5. Store this token securely as an environment variable:
   ```bash
   export DISCORD_BOT_TOKEN="your-bot-token-here"
   ```
   Or add it to your environment variables permanently

### 2. Configure Bot Permissions

1. In the "Bot" tab, under "Privileged Gateway Intents", enable:
   - Message Content Intent
   - Server Members Intent
   - Presence Intent

2. Under "Bot Permissions", ensure the following permissions are enabled:
   - Send Messages
   - Embed Links
   - Attach Files
   - Read Message History
   - Use External Emojis

### 3. Invite Bot to Your Server

1. Go to the "OAuth2" tab, then "URL Generator"
2. Select "bot" under "SCOPES"
3. Select the permissions mentioned above
4. Copy the generated URL and open it in a browser
5. Select your server and authorize the bot

### 4. Get Channel IDs

1. In Discord, enable Developer Mode in Settings > Advanced
2. Right-click on any channel and select "Copy ID"
3. Add these channel IDs to your config.yaml file under the appropriate audience

## Setting Up Telegram Bot

To use Telegram notifications, you need to create a Telegram bot and get the chat IDs. Follow these steps:

### 1. Create a Telegram Bot

1. Open Telegram and search for the "BotFather" (@BotFather)
2. Start a chat with BotFather and send the command `/newbot`
3. Follow the instructions to create a new bot:
   - Provide a name for your bot (e.g., "ZetaChain Notifications")
   - Provide a username for your bot (must end with "bot", e.g., "zetachain_notifications_bot")
4. BotFather will provide a token for your new bot
5. Store this token securely as an environment variable:
   ```bash
   export TELEGRAM_BOT_TOKEN="your-bot-token-here"
   ```
   Or add it to your environment variables permanently

### 2. Add Bot to Groups/Channels

1. Create a new group or channel in Telegram (or use an existing one)
2. Add your bot to the group/channel as an administrator
3. For channels, give the bot "Post Messages" permission
4. For groups, no special permissions are needed beyond being a member

### 3. Get Chat IDs

For groups and channels, you need to get the chat ID to send messages. There are several ways to do this:

#### Method 1: Using the Telegram API

1. Add your bot to the group/channel
2. Send a message in the group/channel
3. Open this URL in your browser (replace YOUR_BOT_TOKEN with your actual bot token):
   ```
   https://api.telegram.org/botYOUR_BOT_TOKEN/getUpdates
   ```
4. Look for the "chat" object in the response and find the "id" field
5. For groups, the ID will be negative (e.g., -1001234567890)

#### Method 2: Using a Chat ID Bot

1. Add the "Get My ID" bot (@getmyid_bot) to your group
2. The bot will display the chat ID

### 4. Configure Your Application

1. Add the chat IDs to your config.yaml file under the appropriate audience
2. Make sure the TELEGRAM_BOT_TOKEN environment variable is set

## Usage

```bash
# Run with default configuration
go run main.go

# Run with additional configuration file
go run main.go --config additional-config.yaml
```

## Development

### Project Structure

- `configs/`: Configuration files
- `internal/`: Internal packages
  - `comms/`: Communication services
  - `config/`: Configuration handling
  - `events/`: Event processing
  - `storage/`: Storage services
- `pkg/`: Public packages
  - `discord/`: Discord client
  - `slack/`: Slack client
  - `telegram/`: Telegram client
  - `zetachain/`: ZetaChain API client

### Adding a New Notification Channel

1. Create a new package in `pkg/` for the channel
2. Implement the client with appropriate message formatting
3. Update the `NotificationService` in `internal/comms/notifications.go`
4. Update the configuration structure in `internal/config/config.go`

## TODOs

- [x] Slack integration
- [x] Discord integration
- [x] Telegram integration
- [x] Improved message formatting
- [x] Run for all networks
- [x] update config , add bot tokens to single config and maybe split in structs?
- [ ] Ask to refactor notifications to make it common
- [ ] broadcast messages
- [ ] Pagination for proposal fetching
- [ ] Comprehensive tests
- [ ] Docs
- [x] loook at type MsgSoftwareUpgrade, in config or generic check
- [ ] test upgrade by submtted proposal
