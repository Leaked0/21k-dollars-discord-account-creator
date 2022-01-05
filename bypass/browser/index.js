const Websocket = require('ws');
const { EventEmitter } = require('events');
const { execSync } = require('child_process');

const app = (require('express'))();

class SocketListener extends EventEmitter {
    constructor() {
        super();

        this.server = new Websocket.Server({
            port: 3200
        });

        this.client = undefined;
    }

    async start() {
        this.server.on("connection", (socket) => {
            this.client = socket;
            this.client.on('message', (data) => this.emit('resolve', data));

            this.emit('ready');
        });
    }

    send(data) {
        this.client.send(JSON.stringify(data));
    }
}

const Client = new SocketListener();

app.get('/n', async (req, res) => {
    await Client.send({
        solve: req.query.req
    });

   Client.once('resolve', async (data) => {
       res.send(JSON.parse(data).token);
   });
});

app.listen(3030, async () => {
    await Client.start();
});
