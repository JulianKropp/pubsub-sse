// TODO:
// - [ ] What if the main tab sse connection fails?
// - [ ] What if connection_id changes or there are multiple connection_id?
// - [ ] if error connecting to broadcast maybe wrong connection_id becomming master. Then there are two masters.


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
        this.status = `disconnected`; // connecting, connected, or disconnected
        this.connection_type = `sse`; // sse or broadcastchannel
        this.channel = null;
        this.evtSource = null;
        this.topics = new Map(); // Stores Topic objects

        this.onConnected = null;
        this.onDisconnected = null;
        this.onError = null;
        this.onNewTopic = null;
        this.onRemovedTopic = null;

        let thistab = new Tab();
        this.tab = thistab;
        this.tabs = new Map(); // Stores Tab objects
        this.tabs.set(this.tab.id, thistab);
        this.masterTab = null;

        this.tabChannel = new BroadcastChannel('tab_communication');

        this.unknownInstances = new Map();
        this.removeInstanceAfterXAttempts = 3;

        // Timeout and intervals
        this.pingInterval = 100;
        this.checkTabsInterval = 500;
        this.timeout = 1000;

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

        // Find tab
        const tab = this.tabs.get(tabID);
        if (tab) {
            data.tabID = tab.id;
            data.instance_id = this.tab.instance_id;
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
        if (!this.tabs.has(from)) {
            this.tabs.set(from, new Tab(from));
        }
        if (!this.tabs.has(data.tabID)) {
            this.tabs.set(data.tabID, new Tab(data.tabID));
        }

        // Get data tab
        const dataTab = this.tabs.get(data.tabID);
        if (dataTab === this.masterTab && data.connection_id !== dataTab.connection_id) {
            console.log(`Master tab connection_id changed. Reconnecting.`);
            this.changeConnection_id();
        }
        dataTab.instance_id = data.instance_id;
        dataTab.connection_id = data.connection_id;

        switch (type) {
            case 'new':
                this.broadcast('ack', this.tab.id);
                this.electMaster();
                break;
            case 'ack':
            case 'ping':
                this.tabs.get(from).resetTimer();
                break;
            case 'master':
                if (this.isMaster() && data.tabID > this.tab.id) {
                    this.electMaster(); // Elect self if older (lower ID)
                } else if (data !== this.tab.id) {
                    this.setMaster(data.tabID);
                }
                break;
        }
    }

    checkTabs() {
        const now = Date.now();
        for (let [tabId, tab] of this.tabs.entries()) {
            if (this.tab.id !== tabId) {
                if (now - tab.lastMessage > this.timeout) {
                    console.log(`Removing inactive tab ${tabId}`);

                    if (this.masterTab) {
                        if (this.isMaster() || tab.id === this.masterTab.id) {
                            this.removeInstance(tab.instance_id);
                        }
                    }
                    this.tabs.delete(tabId);
                    this.electMaster();
                }
            }
        }
    }

    electMaster() {
        // Ensure current tab is in the list
        if (!this.tabs.has(this.tab.id)) {
            this.tabs.set(this.tab.id, this.tab);
        }

        // Find the tab with the lowest ID to elect as master
        const lowestId = Math.min(this.tab.id, ...Array.from(this.tabs.keys()).map(Number));
        if (lowestId === this.tab.id) {
            this.broadcast('master', this.tab.id);
            this.setMaster(this.tab.id);
        } else {
            // Set the master tab if another tab is elected master
            this.setMaster(lowestId);
        }
    }

    // Returns true if this tab is the master tab
    isMaster() {
        return this.masterTab && this.masterTab.id === this.tab.id;
    }

    setMaster(masterID) {
        const wasMaster = this.isMaster();
        if (!this.masterTab || this.masterTab.id !== masterID) {
            this.masterTab = this.tabs.get(masterID);
            console.log(`Master tab is now ${masterID}`);
        }

        const isNowMaster = this.isMaster();
        if (wasMaster !== isNowMaster) {
            this.updateStatus();
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

    removeInstance(instance_id) {
        if (instance_id) {
            const xhr = new XMLHttpRequest();
            xhr.open(`GET`, `${this.url}/remove/user?instance_id=${instance_id}`);
            xhr.send();
            xhr.onload = () => {
                if (xhr.status !== 200) {
                    console.log(`Error removing instance.`);
                    return;
                }
                console.log(`Instance removed.`);
            };
            xhr.onerror = () => {
                console.log('Failed to remove instance due to network error or server unavailability.');
            };
        }
    }

    open() {
        if (this.status === 'connecting') {
            return;
        }
        this.status = 'connecting';
        this.tab.instance_id = null;
        this.tab.connection_id = null;

        const waitForMaster = () => {
            // Wait until this.masterTab isn't null
            if (!this.masterTab) {
                setTimeout(() => {
                    console.log(`Waiting for master tab: ${JSON.stringify(this.masterTab)}`);
                    waitForMaster();
                }, 100);
                return;
            } else {
                this.setConnectionType();
                this.registerNewInstance();
            }
        };

        // Call the waitForMaster function to start waiting
        waitForMaster();
    }

    registerNewInstance() {
        const xhr = new XMLHttpRequest();
        if (this.masterTab.connection_id) {
            // This will register a new instance and will send all the data to the connection_id which is normally another tab (instance)
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

    changeConnection_id() {
        if (this.tab.instance_id && this.masterTab.connection_id) {
            const xhr = new XMLHttpRequest();
            xhr.open(`GET`, `${this.url}/update/user?instance_id=${this.tab.instance_id}&connection_id=${this.masterTab.connection_id}`);
            xhr.send();
            xhr.onload = () => {
                if (xhr.status !== 200) {
                    console.log(`Error changing connection_id.`);
                    return;
                }
                console.log(`Connection_id changed.`);
                this.tab.connection_id = this.masterTab.connection_id;
                this.startConnection();
            };
            xhr.onerror = () => {
                console.log('Failed to change connection_id due to network error or server unavailability.');
                if (this.onError) {
                    this.onError();
                }
            };
        }
    }

    startConnection() {
        // Close existing connection
        if (this.evtSource) {
            this.evtSource.close();
        }

        // Remove broadcastchannel
        if (this.channel != null) {
            this.channel.close();
        }
        this.channel = new BroadcastChannel(`pubsubsse_communication_${this.tab.connection_id}`);
        this.channel.onmessage = null;

        // Set status to connected
        this.status = `connected`;

        // Open new connection based on connection type
        if (this.connection_type === `sse`) {
            this.connectWithSSE(); // Open a new SSE connection
        } else if (this.connection_type === `broadcastchannel`) {
            if (this.masterTab.connection_id !== this.tab.connection_id) {
                this.changeConnection_id();
            }
            this.connectWithBroadcastChannel(); // Listen for messages on the BroadcastChannel
        }
    }

    // Open a new SSE connection
    connectWithSSE() {
        // Close existing connection
        if (this.evtSource) {
            this.evtSource.close();
        }

        this.evtSource = new EventSource(`${this.url}/event?instance_id=${this.tab.instance_id}`);

        // On open event
        this.evtSource.onopen = () => {
            console.log(`Connection to server opened.`);
            if (this.onConnected) {
                this.onConnected();
            }

            // Send onopen event to the broadcastchannel
            this.channel.postMessage({ type: `event_open` });
        };

        // On message event
        this.evtSource.onmessage = (e) => {
            const data = JSON.parse(e.data);

            console.log(`Received sse message: ` + e.data);

            // Publish data in channel
            this.channel.postMessage({ type: `data`, data });

            // Handle the data
            this.handleData(data);
        };

        // On error event
        this.evtSource.onerror = () => {
            console.log(`EventSource failed.`);
            if (this.onError) {
                this.onError();
            }

            // Send onerror event to the broadcastchannel
            this.channel.postMessage({ type: `event_error` });

            // Reconnect
            this.open();
        };

        // On close event
        this.evtSource.onclose = () => {
            console.log(`Connection to server closed.`);
            if (this.onDisconnected) {
                this.onDisconnected();
            }

            // Send onclose event to the broadcastchannel
            this.channel.postMessage({ type: `event_close` });
        };
    }

    // Listen for messages on the BroadcastChannel
    connectWithBroadcastChannel() {
        // On message event
        this.channel.onmessage = (event) => {
            // On message event
            if (event.data.type === `data`) {
                const data = event.data.data;

                console.log(`Received broadcastchannel message: ` + JSON.stringify(data));

                // Handle the data
                this.handleData(data);
            }

            // On open event
            if (event.data.type === `event_open`) {
                console.log(`Connection to server opened.`);
                if (this.onConnected) {
                    this.onConnected();
                }
            }

            // On error event
            if (event.data.type === `event_error`) {
                console.log(`EventSource failed.`);
                if (this.onError) {
                    this.onError();
                }
            }

            // On close event
            if (event.data.type === `event_close`) {
                console.log(`Connection to server closed.`);
                if (this.onDisconnected) {
                    this.onDisconnected();
                }
            }
        };
    }

    handleData(data) {
        // Parse data (assuming it's correctly formatted)
        const instances = data.instances;

        // Get data with this instance_id
        // Example data: {`instances`:[{`id`:`I-ba6b975b-4b18-40d1-88c5-ee947f1888d9`,`data`:{`sys`:null,`updates`:[{`topic`:`T-bc57baf3-4c76-4ae9-a8dd-0f1726f7ea47`,`data`:`DATAAAAAA`}]}}]}
        const instanceData = instances.find(instance => instance.id === this.tab.instance_id);
        if (!instanceData) {
            return;
        }

        // If instance doesn't exist 10 times, remove it
        if (this.isMaster()) {
            for (const instance of instances) {
                // Check if the instance exists in tabs
                const instanceExists = Array.from(this.tabs.values()).some(tab => tab.instance_id === instance.id);

                // If instance not in tabs.instance_id, add instance id to unknownInstances and increment counter
                if (!instanceExists) {
                    console.log(`Instance not found: ` + instance.id);
                    const attempts = this.unknownInstances.get(instance.id) || 0;
                    this.unknownInstances.set(instance.id, attempts + 1);

                    // If counter > removeInstanceAfterXAttempts, remove instance
                    if (attempts + 1 > this.removeInstanceAfterXAttempts) {
                        this.removeInstance(instance.id);
                        this.unknownInstances.set(instance.id, 0);
                    }
                }
            }
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
                    const topic = this.topics.get(topicId);
                    if (topic) {
                        if (topic.subscribed) {
                            topic.onUnsubscribed?.(); // Call the onUnsubscribed event if defined
                        }
                        this.topics.delete(topicId);
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
                    const topic = this.topics.get(topicInfo.ID);
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
            const topic = this.topics.get(update.topic);
            if (topic) {
                topic.onUpdate?.(update.data); // Call the onUpdate event if defined
            }
        });
    }

    ensureTopic(ID, type) {
        let topic; // Define a local variable for the topic
        if (!this.topics.has(ID)) {
            console.log(`New topic: ` + ID);
            topic = new Topic(ID, type); // Assign the new Topic to the local variable
            this.topics.set(ID, topic); // Store the Topic in the topics Map
            this.onNewTopic(topic); // Notify client of new topic using the local variable
        } else {
            topic = this.topics.get(ID); // If the topic already exists, assign it to the local variable
        }
        return topic;
    }
}
