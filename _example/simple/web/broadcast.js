class Tab {
    constructor() {
        this.id = Date.now() + Math.random(); // Unique identifier for each tab, lower is older
        this.isMaster = false;
        this.channel = new BroadcastChannel('tab_communication');
        this.tabs = {}; // Keeps track of other tabs

        this.channel.onmessage = this.handleMessage.bind(this);
        window.addEventListener('unload', () => this.broadcast('left', this.id));

        // Broadcast presence to other tabs and request current status
        this.broadcast('new', this.id);
        setTimeout(() => this.electMaster(), 100); // Wait for responses
    }

    broadcast(type, data) {
        this.channel.postMessage({ type, data, from: this.id });
    }

    handleMessage(event) {
        const { type, data, from } = event.data;
        if (from === this.id) return; // Ignore self-sent messages

        switch (type) {
            case 'new':
                this.tabs[from] = data; // Acknowledge new tab
                this.broadcast('ack', this.id); // Acknowledge back
                this.electMaster();
                break;
            case 'ack':
                this.tabs[from] = data; // Register acknowledged tab
                break;
            case 'left':
                delete this.tabs[from]; // Remove tab from list
                this.electMaster();
                break;
            case 'master':
                if (this.isMaster && data > this.id) {
                    this.electMaster(); // Elect self if older (lower ID)
                } else if (data !== this.id) {
                    this.isMaster = false;
                    this.updateStatus();
                }
                break;
        }
    }

    electMaster() {
        const lowestId = Math.min(this.id, ...Object.keys(this.tabs).map(key => parseFloat(key)));
        if (lowestId === this.id) {
            this.isMaster = true;
            this.broadcast('master', this.id);
        } else {
            this.isMaster = false;
        }
        this.updateStatus();
    }

    updateStatus() {
        console.log('Status:', this.isMaster ? 'Master' : 'Slave');
        const statusElement = document.getElementById('status');
        statusElement.textContent = this.isMaster ? 'Master' : 'Slave';
    }
}