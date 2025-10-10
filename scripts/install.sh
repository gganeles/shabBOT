#!/bin/sh
set -e

# Determine repository root (script lives in ./scripts)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"



# initial install
npm install

# manually install bot-wa-baileys without dependencies
cd node_modules
git clone https://github.com/andresayac/bot-wa-baileys

# fix dependencies
cd bot-wa-baileys
npm install @whiskeysockets/baileys@6.7.18
npm install node-cache
npm install @naanzitos/baileys-make-in-memory-store@1.0.5

# manually build bot-wa-baileys — continue even if build fails
npm run build > /dev/null 2>&1 || echo "Warning: 'npm run build' had errors, continuing."


# Edit specific lines in ./lib/baileys.js:

# Pass upwards WAMessage
echo "Patching lib/baileys.js to pass upwards WAMessage (creating backup lib/baileys.js.bak)..."
# Insert the exact line with leading pluses (escaped for double-quoted shell string)
sed -zi "s/payload.from = utils_1.default.formatPhone(payload.from, this.plugin);[[:space:]]\\+this.emit('message', payload);/payload.from = utils_1.default.formatPhone(payload.from, this.plugin);\\n  payload.messagesObj = messages;\\nthis.emit('message', payload);/" lib/baileys.js


# Turn off printQRInTerminal (depreciated)
echo "Patching lib/baileys.js to turn off printQRInTerminal (creating backup lib/baileys.js.bak)..."
sed -zi 's/printQRInTerminal:[^\n]\+/printQRInTerminal: false,/' lib/baileys.js

# add phone numbers to the command groupMetadata in @whiskeysockets/baileys
echo "Patching @whiskeysockets/baileys/lib/Socket/groups.js to add phone numbers (creating backup groups.js.bak)..."
sed -zi 's/admin: (addrs.type || null),/admin: (addrs.type || null),\n number: (attrs.phone_number || null),/' node_modules/@whiskeysockets/baileys/lib/Socket/groups.js

