import WA from 'baileys';
import P from "pino";
import { NodeCache } from "node-cache";

class Client extends EventEmitter {
    constructor() {
        super();
        const user = {
            logger: P({ level: 'silent' }),
            auth: WA.useMultiFileAuthState('./session_data'),
            browser: ['Client', 'Safari', '1.0.0'],
            cachedGroupMetadata: async (jid) => groupCache.get(jid)
        }

        this.sock = WA.makeWASocket(user);

        this.sock.ev.on('connection.update', (update) => {
            const { connection, lastDisconnect, qr } = update;
            if (qr) {
                this.emit('qr', qr);
            }
            if (connection === 'close') {
                const shouldReconnect = lastDisconnect.error?.output?.statusCode !== DisconnectReason.loggedOut;
                if (shouldReconnect) {
                    this.sock = WA.makeWASocket();
                }
            }
        });
    }