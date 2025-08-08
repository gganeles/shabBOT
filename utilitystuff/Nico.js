//const {MessageMedia} = require("whatsapp-web.js")
//const fs = require("fs")

function dontBeAPussy(client, chat) {
    //file = fs.readFileSync("utilitystuff/nico.opus", { encoding: 'base64' })
    //const vm = new MessageMedia("audio/ogg", file)
    client.sendAudio(chat, "utilitystuff/nico.opus")
}

module.exports = {dontBeAPussy}


