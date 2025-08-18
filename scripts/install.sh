#!/bin/sh
set -e


# initial install
npm install

# manually install bot-wa-baileys without dependencies
cd node_modules
git clone https://github.com/andresayac/bot-wa-baileys

# fix dependencies
cd bot-wa-baileys
npm install @whiskeysockets/baileys@latest 
npm install @naanzitos/baileys-make-in-memory-store

# manually build bot-wa-baileys
tsc


# Edit specific lines in ./lib/baileys.js:

# Pass upwards WAMessage
sed -zi 's/payload.from = utils_1.default.formatPhone(payload.from, this.plugin);[[:space:]]\+this.emit('message', payload);/payload.from = utils_1.default.formatPhone(payload.from, this.plugin);\npayload.messagesObj = messages;\nthis.emit('message', payload);/' lib/baileys.js

# Turn off printQRInTerminal (depreciated)
sed -zi 's/printQRInTerminal:[^\n]\+/printQRInTerminal: false;/' lib/baileys.js

# add phone numbers to the command groupMetadata in @whiskeysockets/baileys