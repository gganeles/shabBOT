import { BaileysClass } from 'bot-wa-baileys';
import EventEmitter from 'events';


class Message {
    constructor(baileysMessage, client) {
        this.from = baileysMessage.key.remoteJid;
        this.body = baileysMessage.message?.conversation ||
            baileysMessage.message?.extendedTextMessage?.text || '';
        this.client = client;
        this.pushName = baileysMessage.pushName || '';
        this.type = baileysMessage.type || 'uu';
        this.WAMessage = baileysMessage.messagesObj[0];
        this.messageObj = baileysMessage;
        this.isGroup = baileysMessage.key.remoteJid.endsWith('@g.us');
        this.key = baileysMessage.key;
    }

    async reply(text) {
        return this.client.baileys.sendMessage(this.from, text, {
            quoted: this.WAMessage,
            options: {}
        });
    }
}

class Client extends EventEmitter {
    constructor() {
        super();
        this.sock = null;
        this.baileys = new BaileysClass({ debug: false });
        // Set up event listeners
        this.baileys.on('qr', (qr) => {
            this.emit('qr', qr);
        });

        this.baileys.on('auth_failure', (error) => {
            console.error('Authentication failed:', error);
            this.emit('auth_failure', error);
        });

        this.baileys.on('ready', () => {
            this.emit('ready');
        });

        this.baileys.on('message', (message) => {
            //console.log('Received message:', JSON.stringify(message, null, 2));
            if (message.body) {
                const messageObj = new Message(message, this);
                this.emit('message', messageObj);
            }
        });

    }

    async sendMessage(to, message) {
        const response = await this.baileys.sendMessage(to, message, {options:{}});
        return response;
    }

    async sendAudio(number, audioUrl) {
        this.baileys.sendAudio(number, audioUrl);
    }

    async getGroupMetadata(jid) {
        const groups = await this.baileys.vendor.groupMetadata(jid);
        return groups;
    }
}

export { Client };