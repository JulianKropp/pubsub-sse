class Tab {
    constructor() {
        this.id = Date.now() + Math.random(); // Unique identifier for each tab, lower is older
        this.isMaster = false;
        this.channel = new BroadcastChannel('tab_communication');
        this.tabs = {}; // Keeps track of other tabs

        // Timeout and intervals
        this.pingInterval = 1000;
        this.checkTabsInterval = 1000;
        this.timeout = 3000;

        this.channel.onmessage = this.handleMessage.bind(this);

        // Broadcast presence to other tabs and request current status
        this.broadcast('new', this.id);
        setTimeout(() => this.electMaster(), 100); // Wait for responses

        this.pingInterval = setInterval(() => this.broadcast('ping', this.id), this.pingInterval);
        this.checkTabsInterval = setInterval(() => this.checkTabs(), this.checkTabsInterval);
    }

    broadcast(type, data) {
        this.channel.postMessage({ type, data, from: this.id });
    }

    handleMessage(event) {
        const { type, data, from } = event.data;
        if (from === this.id) return; // Ignore self-sent messages

        switch (type) {
            case 'new':
                this.tabs[from] = Date.now(); // Acknowledge new tab and set last ping time
                this.broadcast('ack', this.id); // Acknowledge back
                this.electMaster();
                break;
            case 'ack':
                this.tabs[from] = Date.now(); // Register acknowledged tab and set last ping time
                break;
            case 'ping':
                this.tabs[from] = Date.now(); // Update last ping time for this tab
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

    checkTabs() {
        const now = Date.now();
        Object.keys(this.tabs).forEach(tabId => {
            if (now - this.tabs[tabId] > this.timeout) { // Remove tab if last ping was more than 5 seconds ago
                delete this.tabs[tabId];
                this.electMaster();
            }
        });
    }

    electMaster() {
        if (Object.keys(this.tabs).length === 0 || !this.tabs[this.id]) {
            this.tabs[this.id] = Date.now(); // Ensure current tab is in list
        }
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
        if (statusElement) {
            statusElement.textContent = this.isMaster ? 'Master' : 'Slave';
        }
    }
}
