# Environment Variables

## Config

 - `PORT` (default: `1323`) - Port to listen for incoming connections
 - `TIMEOUT` (default: `60s`) - Timeout for outgoing http requests
 - `BOT_UA` (default: `Mozilla/5.0 (compatible; SummalyBot/0.0.1; +https://github.com/yulog/go-summaly)`) - BotUA
 - `NON_BOT_UA` (default: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML`) - NonBotUA
 - `REQUIRE_NON_BOT_UA_FILE` (default: `./nonbot.txt`) - RequireNonBotUAFile
 - `REQUIRE_NON_BOT_UA` (comma-separated, expand, from-file, default: `${REQUIRE_NON_BOT_UA_FILE}`) - RequireNonBotUA
 - `HIDE_BANNER` (default: `false`) - HideBanner to hide startup banner
 - `ALLOW_PRIVATE_IP` (default: `false`) - AllowPrivateIP to connect private ip for test

