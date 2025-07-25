
# Setting Up Discord Bot

To use Discord notifications, you need to create a Discord bot and add it to your server. Follow these steps:

## 1. Create a Discord Application

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and give it a name (e.g., "ZetaChain Notifications")
3. Navigate to the "Bot" tab and click "Add Bot"
4. Under the "TOKEN" section, click "Copy" to get your bot token
5. Store this token securely as an environment variable:
   ```bash
   export DISCORD_BOT_TOKEN="your-bot-token-here"
   ```
   Or add it to your environment variables permanently

## 2. Configure Bot Permissions

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

## 3. Invite Bot to Your Server

1. Go to the "OAuth2" tab, then "URL Generator"
2. Select "bot" under "SCOPES"
3. Select the permissions mentioned above
4. Copy the generated URL and open it in a browser
5. Select your server and authorize the bot

## 4. Get Channel IDs

1. In Discord, enable Developer Mode in Settings > Advanced
2. Right-click on any channel and select "Copy ID"
3. Add these channel IDs to your config.yaml file under the appropriate audience
