
# Setting Up Telegram Bot

To use Telegram notifications, you need to create a Telegram bot and get the chat IDs. Follow these steps:

## 1. Create a Telegram Bot

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

## 2. Add Bot to Groups/Channels

1. Create a new group or channel in Telegram (or use an existing one)
2. Add your bot to the group/channel as an administrator
3. For channels, give the bot "Post Messages" permission
4. For groups, no special permissions are needed beyond being a member

## 3. Get Chat IDs

For groups and channels, you need to get the chat ID to send messages. There are several ways to do this:

### Method 1: Using the Telegram API

1. Add your bot to the group/channel
2. Send a message in the group/channel
3. Open this URL in your browser (replace YOUR_BOT_TOKEN with your actual bot token):
   ```
   https://api.telegram.org/botYOUR_BOT_TOKEN/getUpdates
   ```
4. Look for the "chat" object in the response and find the "id" field
5. For groups, the ID will be negative (e.g., -1001234567890)

### Method 2: Using a Chat ID Bot

1. Add the "Get My ID" bot (@getmyid_bot) to your group
2. The bot will display the chat ID

## 4. Configure Your Application

1. Add the chat IDs to your config.yaml file under the appropriate audience
2. Make sure the TELEGRAM_BOT_TOKEN environment variable is set
