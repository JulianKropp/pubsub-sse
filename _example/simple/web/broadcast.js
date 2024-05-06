class TabManager {
    constructor() {
        let thistab = new Tab()
        this.tab = thistab;
        this.tabs = [];
        this.tabs[this.tab.id] = thistab;
        this.masterTab = null;

        this.channel = new BroadcastChannel('tab_communication');
        console.log(`TabManager initialized with tab ID ${this.tab.id}`);

        // Timeout and intervals
        this.pingInterval = 100;
        this.checkTabsInterval = 500;
        this.timeout = 1000;

        this.channel.onmessage = this.handleMessage.bind(this);

        // Broadcast presence to other tabs and request current status
        this.broadcast('new', this.tab.id);
        setTimeout(() => {
            console.log('Electing master after initial timeout');
            this.electMaster();
        }, this.timeout); // Wait for responses

        this.pingIntervalId = setInterval(() => {
            console.log('Sending ping');
            this.broadcast('ping', this.tab.id);
        }, this.pingInterval);

        this.checkTabsIntervalId = setInterval(() => {
            console.log('Checking tabs');
            this.checkTabs();
        }, this.checkTabsInterval);
    }

    broadcast(type, data) {
        console.log(`Broadcasting message type: ${type}, data: ${data}`);
        this.channel.postMessage({ type, data, from: this.tab.id });
    }

    handleMessage(event) {
        const { type, data, from } = event.data;
        console.log(`Received message from ${from}: ${type}`);
        if (from === this.tab.id) {
            console.log('Ignoring self-sent message');
            return; // Ignore self-sent messages
        }

        // Create tab if it doesn't exist
        if (!this.tabs[from]) {
            this.tabs[from] = new Tab(from);
            console.log(`New tab created with ID ${from}`);
        }

        switch (type) {
            case 'new':
                this.broadcast('ack', this.tab.id);
                this.electMaster();
                break;
            case 'ack':
            case 'ping':
                this.tabs[from].resetTimer();
                console.log(`Timer reset for tab ${from}`);
                break;
            case 'master':
                if (this.tab.isMaster && data > this.tab.id) {
                    console.log('Higher ID master detected, re-electing');
                    this.electMaster(); // Elect self if older (lower ID)
                } else if (data !== this.tab.id) {
                    this.setMaster(data);
                }
                break;
        }
    }

    checkTabs() {
        const now = Date.now();
        Object.keys(this.tabs).forEach(tabId => {
            if (now - this.tabs[tabId].lastMessage > this.timeout) {
                console.log(`Removing inactive tab ${tabId}`);
                if (this.tab.id !== tabId) {
                    delete this.tabs[tabId];
                }
                this.electMaster();
            }
        });
    }

    electMaster() {
        console.log('Electing master tab');
        if (Object.keys(this.tabs).length === 0 || !this.tabs[this.tab.id]) {
            this.tabs[this.tab.id] = this.tab; // Ensure current tab is in list
        }
        const lowestId = Math.min(this.tab.id, ...Object.keys(this.tabs).map(key => parseFloat(key)));
        console.log(`Lowest ID: ${lowestId}`);
        if (lowestId === this.tab.id) {
            this.broadcast('master', this.tab.id);
            this.setMaster(this.tab.id);
        }
    }

    isMaster() {
        return this.masterTab && this.masterTab.id === this.tab.id;
    }

    setMaster(master) {
        const oldIsMaster = this.isMaster();
        if (!this.masterTab || this.masterTab.id !== master) {
            this.masterTab = this.tabs[master];
            console.log(`Master set to tab ID ${master}`);

            if (oldIsMaster !== this.isMaster()) {
                this.updateStatus();
            }
        }
    }

    updateStatus() {
        console.log('Status:', this.isMaster() ? 'Master' : 'Slave');
        const statusElement = document.getElementById('status');
        if (statusElement) {
            statusElement.textContent = this.isMaster() ? 'Master' : 'Slave';
        }
    }
}

class Tab {
    constructor(id = null) {
        if (id === null) {
            this.id = Date.now() + Math.random(); // Unique identifier for each tab, lower is older
        } else {
            this.id = id;
        }

        this.lastMessage = Date.now();
        console.log(`Tab created with ID ${this.id}`);
    }

    resetTimer() {
        console.log(`Resetting last message time for tab ${this.id}`);
        this.lastMessage = Date.now();
    }
}
