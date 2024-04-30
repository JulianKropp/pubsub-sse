// TODO:
// - [ ] What if main tab closes? Fallback to SSE?
// - [ ] What if the main tab reloads. It will get a new connection_id and all outher tabs will timeout and will only receve the data of the new connection
// - [ ] What if second tab closes? How to handle the instance which should not receve any messages anymore, but the main tab still does for the second tab?

class Topic {
    constructor(id, type) {
        this.id = id;
        this.type = type;
        this.subscribed = false;
        this.onSubscribed = null;
        this.onUnsubscribed = null;
        this.onUpdate = null;
    }

    // Additional methods for topic-related functionalities
}

class Tab {
    constructor(id = null) {
        if (id === null) {
            this.id = Date.now() + Math.random(); // Unique identifier for each tab, lower is older
        } else {
            this.id = id;
        }

        this.lastMessage = Date.now();

        this.instance_id = null;
        this.connection_id = null;

        console.log(`New tab: ${this.id}`);

    }

    resetTimer() {
        this.lastMessage = Date.now();
    }
}

class PubSubSSE {
    constructor(url = `http://localhost`) {
        this.url = url;
        this.status = `disconnected`; // connecting, connected or disconnected
        this.connection_type = `sse`; // sse or broadcastchannel
        this.channel = new BroadcastChannel('pubsubsse_communication');
        this.evtSource = null;
        this.topics = {}; // Stores Topic objects

        this.onConnected = null;
        this.onDisconnected = null;
        this.onError = null;
        this.onNewTopic = null;
        this.onRemovedTopic = null;


        let thistab = new Tab()
        this.tab = thistab;
        this.tabs = [];
        this.tabs[this.tab.id] = thistab;
        this.masterTab = null;

        this.tabChannel = new BroadcastChannel('tab_communication');

        // Timeout and intervals
        this.pingInterval = 100;
        this.checkTabsInterval = 100;
        this.timeout = 300;

        this.tabChannel.onmessage = this.handleMessage.bind(this);

        // Broadcast presence to other tabs and request current status
        this.broadcast('new', this.tab.id);
        setTimeout(() => {
            console.log('Electing master after initial timeout');
            this.electMaster();
        }, this.timeout); // Wait for responses

        this.pingIntervalId = setInterval(() => {
            this.broadcast('ping', this.tab.id);
        }, this.pingInterval);

        this.checkTabsIntervalId = setInterval(() => {
            this.checkTabs();
        }, this.checkTabsInterval);
    }

    broadcast(type, tabID) {
        let data = { tabID: null, instance_id: null, connection_id: null };

        // find tab
        const tab = this.tabs[tabID];
        if (tab) {
            data.tabID = tab.id;
            data.instance_id =  this.tab.instance_id;
            data.connection_id = this.tab.connection_id;
        } else {
            console.log(`Tab ${tabID} not found`);
            data.tabID = tabID;
        }

        this.tabChannel.postMessage({ type, data, from: this.tab.id });
    }

    handleMessage(event) {
        const { type, data, from } = event.data;
        if (from === this.tab.id) {
            return; // Ignore self-sent messages
        }

        // Create tab if it doesn't exist
        if (!this.tabs[from]) {
            this.tabs[from] = new Tab(from);
        }
        if (!this.tabs[data.tabID]) {
            this.tabs[data.tabID] = new Tab(data.tabID);
        }

        // Get data tab
        const dataTab = this.tabs[data.tabID];
        dataTab.instance_id = data.instance_id;
        dataTab.connection_id = data.connection_id;

        switch (type) {
            case 'new':
                this.broadcast('ack', this.tab.id);
                this.electMaster();
                break;
            case 'ack':
            case 'ping':
                this.tabs[from].resetTimer();
                break;
            case 'master':
                if (this.tab.isMaster && data.tabID > this.tab.id) {
                    this.electMaster(); // Elect self if older (lower ID)
                } else if (data !== this.tab.id) {
                    this.setMaster(data.tabID);
                }
                break;
        }
    }

    checkTabs() {
        const now = Date.now();
        Object.keys(this.tabs).forEach(tabId => {
            if (this.tab.id != tabId) {
                if (now - this.tabs[tabId].lastMessage > this.timeout) {
                    console.log(`Removing inactive tab ${tabId}`);
                    delete this.tabs[tabId];
                    this.electMaster();
                }
            }
        });
    }

    electMaster() {
        if (Object.keys(this.tabs).length === 0 || !this.tabs[this.tab.id]) {
            this.tabs[this.tab.id] = this.tab; // Ensure current tab is in list
        }
        const lowestId = Math.min(this.tab.id, ...Object.keys(this.tabs).map(key => parseFloat(key)));
        if (lowestId === this.tab.id) {
            this.broadcast('master', this.tab.id);
            this.setMaster(this.tab.id);
        }
    }

    // Returns true if this tab is the master tab
    isMaster() {
        return this.masterTab && this.masterTab.id === this.tab.id;
    }

    setMaster(masterID) {
        const oldIsMaster = this.isMaster();
        if (!this.masterTab || this.masterTab.id !== masterID) {
            this.masterTab = this.tabs[masterID];

            console.log(`Master set to tab ID ${JSON.stringify(masterID)}`);

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

        this.setConnectionType();
        if (this.status === `connected`) {
            this.startConnection();
        }
    }


    setConnectionType() {
        if (this.isMaster()) {
            this.connection_type = `sse`;
        } else {
            this.connection_type = `broadcastchannel`;
        }
    }

    open() {
        this.status = `connecting`;
        this.tab.instance_id = null;
        this.tab.connection_id = null;

        // wait until this.masterTab isnt null
        if (!this.masterTab) {
            setTimeout(() => {
                console.log(`Waiting for master tab: ${JSON.stringify(this.masterTab)}`);
                this.open();
            }, 100);
            return;
        } else {
            this.setConnectionType();
            this.registerNewInstance();
        }
    }

    registerNewInstance() {
        const xhr = new XMLHttpRequest();
        if (this.masterTab.connection_id) {
            // This will register a new instance and will send all the data to the connection_id which is normaly anouther tab (instance)
            // If the connection_id is not valid, the server will return a new connection_id
            xhr.open(`GET`, `${this.url}/add/user?connection_id=${this.masterTab.connection_id}`);
        } else {
            // This will register a new instance and will send all the data to this tab (instance)
            xhr.open(`GET`, `${this.url}/add/user`);
        }
        xhr.send();
        xhr.onload = () => {
            if (xhr.status !== 200) {
                console.log(`Error registering new instance.`);
                return;
            }
            const response = JSON.parse(xhr.responseText);
            this.tab.connection_id = response.connection_id;
            this.tab.instance_id = response.instance_id;
            this.startConnection();
        };
        xhr.onerror = () => {
            console.log('Failed to register new instance due to network error or server unavailability.');
            if (this.onError) {
                this.onError();
            }
        };
    }

    startConnection() {
        // close existing connection
        if (this.evtSource) {
            this.evtSource.close();
        }

        // Remove broadcastchannel
        this.channel.close();
        this.channel = new BroadcastChannel('pubsubsse_communication');
        this.channel.onmessage = null;

        // Set status to connected
        this.status = `connected`;

        // Open new connection based on connection type
        if (this.connection_type === `sse`) {
            this.connectWithSSE(); // Open a new SSE connection
        } else if (this.connection_type === `broadcastchannel`) {
            if (this.masterTab.connection_id === this.tab.connection_id) {
                this.connectWithBroadcastChannel(); // Listen for messages on the BroadcastChannel
            } else {
                // Fall back to SSE if the connection_id is not valid
                this.connectWithSSE();
            }
        }
    }

    // Open a new SSE connection
    connectWithSSE() {
        // close existing connection
        if (this.evtSource) {
            this.evtSource.close();
        }

        this.evtSource = new EventSource(`${this.url}/event?instance_id=${this.tab.instance_id}`);

        // on open event
        this.evtSource.onopen = () => {
            console.log(`Connection to server opened.`);
            if (this.onConnected) {
                this.onConnected();
            }

            // send onopen event to the broadcastchannel
            this.channel.postMessage({ type: `event_open` });
        };

        // on message event
        this.evtSource.onmessage = (e) => {
            const data = JSON.parse(e.data);

            console.log(`Received sse message: ` + e.data);

            // Publish data in channel
            this.channel.postMessage({ type: `data`, data: data });

            // Handle the data
            this.handleData(data);
        };

        // on error event
        this.evtSource.onerror = () => {
            console.log(`EventSource failed.`);
            if (this.onError) {
                this.onError();
            }

            // send onerror event to the broadcastchannel
            this.channel.postMessage({ type: `event_error` });
        };

        // on close event
        this.evtSource.onclose = () => {
            console.log(`Connection to server closed.`);
            if (this.onDisconnected) {
                this.onDisconnected();
            }

            // send onclose event to the broadcastchannel
            this.channel.postMessage({ type: `event_close` });
        }
    }

    // Listen for messages on the BroadcastChannel
    connectWithBroadcastChannel() {
        // on message event
        this.channel.onmessage = (event) => {

            // on message event
            if (event.data.type === `data`) {
                const data = event.data.data;

                console.log(`Received broadcastchannel message: ` + JSON.stringify(data));

                // Handle the data
                this.handleData(data);
            }

            // on open event
            if (event.data.type === `event_open`) {
                console.log(`Connection to server opened.`);
                if (this.onConnected) {
                    this.onConnected();
                }
            }

            // on error event
            if (event.data.type === `event_error`) {
                console.log(`EventSource failed.`);
                if (this.onError) {
                    this.onError();
                }
            }

            // on close event
            if (event.data.type === `event_close`) {
                console.log(`Connection to server closed.`);
                if (this.onDisconnected) {
                    this.onDisconnected();
                }
            }
        }
    }

    handleData(data) {
        // Get data with this instance_id
        // Example data: {`instances`:[{`id`:`I-ba6b975b-4b18-40d1-88c5-ee947f1888d9`,`data`:{`sys`:null,`updates`:[{`topic`:`T-bc57baf3-4c76-4ae9-a8dd-0f1726f7ea47`,`data`:`DATAAAAAA`}]}}]}
        const instanceData = data.instances.find(instance => instance.id ===  this.tab.instance_id);
        if (!instanceData) {
            return;
        }

        this.handleSysMessages(instanceData.sys);
        this.handleUpdateMessages(instanceData.updates);
    }

    handleSysMessages(sysDatas) {
        if (!sysDatas) return;

        sysDatas.forEach(sysData => {
            const type = sysData.type;
            // Types:
            //   `topics`: List of topics
            //   `subscribed`: List of subscribed topics
            //   `unsubscribed`: List of unsubscribed topics

            if (type === `topics`) {
                let removedTopicsList = sysData.list;
                sysData.list.forEach(topicInfo => {
                    this.ensureTopic(topicInfo.id, topicInfo.type);
                    delete removedTopicsList[topicInfo.id];
                });

                // Remove topics that are no longer in the list
                removedTopicsList.forEach(topicId => {
                    const topic = this.topics[topicId];
                    if (topic) {
                        if (topic.subscribed) {
                            topic.onUnsubscribed?.(); // Call the onUnsubscribed event if defined
                        }
                        delete this.topics[topicId];
                        this.onRemovedTopic?.(topic); // Notify client of topic removal
                    }
                });
            } else if (type === `subscribed`) {
                sysData.list.forEach(topicInfo => {
                    const topic = this.ensureTopic(topicInfo.ID, topicInfo.type);
                    topic.onSubscribed?.(); // Call the onSubscribed event if defined
                    topic.subscribed = true; // Mark as subscribed
                });
            } else if (type === `unsubscribed`) {
                sysData.list.forEach(topicInfo => {
                    const topic = this.topics[topicInfo.ID];
                    if (topic) {
                        topic.onUnsubscribed?.(); // Call the onUnsubscribed event if defined
                        topic.subscribed = false; // Mark as unsubscribed
                    }
                });
            }
        });
    }

    handleUpdateMessages(updateData) {
        if (!updateData) return;

        // Handle updates for subscribed topics
        updateData.forEach(update => {
            const topic = this.topics[update.topic];
            if (topic) {
                topic.onUpdate?.(update.data); // Call the onUpdate event if defined
            }
        });
    }

    ensureTopic(ID, type) {
        let topic; // Define a local variable for the topic
        if (!this.topics[ID]) {
            console.log(`New topic: ` + ID);
            topic = new Topic(ID, type); // Assign the new Topic to the local variable
            this.topics[ID] = topic; // Store the Topic in the topics dictionary
            this.onNewTopic?.(topic); // Notify client of new topic using the local variable
        } else {
            topic = this.topics[ID]; // If the topic already exists, assign it to the local variable
        }
        return topic;
    }

}