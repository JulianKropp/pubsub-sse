class Logger {
    constructor(level = 'info') {
        this.levels = {
            'debug': 3,
            'info': 2,
            'warn': 1,
            'error': 0
        };
        this.currentLevel = this.levels[level];
    }

    log(level, message) {
        if (this.levels[level] <= this.currentLevel) {
            console.log(`[${level.toUpperCase()}]: ${message}`);
        }
    }

    debug(message) {
        this.log('debug', message);
    }

    info(message) {
        this.log('info', message);
    }

    warn(message) {
        this.log('warn', message);
    }

    error(message) {
        this.log('error', message);
    }
}

class Topic {
    constructor(id, type, logger) {
        this.id = id;
        this.type = type;
        this.subscribed = false;
        this.onSubscribed = null;
        this.onUnsubscribed = null;
        this.onUpdate = null;
        this.logger = logger;
        this.logger.debug(`Topic created with id: ${id} and type: ${type}`);
    }
}

class Tab {
    constructor(id = null, logger) {
        this.logger = logger || new Logger();
        if (id === null) {
            this.id = Date.now() + Math.random();
        } else {
            this.id = id;
        }
        this.lastMessage = Date.now();
        this.instance_id = null;
        this.connection_id = null;
        this.logger.info(`New tab created: ${this.id}`);
    }

    resetTimer() {
        this.lastMessage = Date.now();
        this.logger.debug(`Timer reset for tab: ${this.id}`);
    }
}

class PubSubSSE {
    constructor(url = `http://localhost`, logLevel = 'debug') {
        this.logger = new Logger(logLevel);
        this.url = url;
        this.status = `disconnected`;
        this.connection_type = `sse`;
        this.channel = null;
        this.evtSource = null;
        this.topics = new Map();
        this.onConnected = null;
        this.onDisconnected = null;
        this.onError = null;
        this.onNewTopic = null;
        this.onRemovedTopic = null;
        let thistab = new Tab(null, this.logger);
        this.tab = thistab;
        this.tabs = new Map();
        this.tabs.set(this.tab.id, thistab);
        this.masterTab = null;
        this.tabChannel = new BroadcastChannel('tab_communication');
        this.unknownInstances = new Map();
        this.removeInstanceAfterXAttempts = 3;
        this.pingInterval = 100;
        this.checkTabsInterval = 500;
        this.timeout = 1000;
        this.tabChannel.onmessage = this.handleMessage.bind(this);
        this.broadcast('new', this.tab.id);
        setTimeout(() => {
            this.logger.info('Electing master after initial timeout');
            this.electMaster();
        }, this.timeout);
        this.pingIntervalId = setInterval(() => {
            this.broadcast('ping', this.tab.id);
        }, this.pingInterval);
        this.checkTabsIntervalId = setInterval(() => {
            this.checkTabs();
        }, this.checkTabsInterval);
    }

    broadcast(type, tabID) {
        let data = { tabID: null, instance_id: null, connection_id: null };
        const tab = this.tabs.get(tabID);
        if (tab) {
            data.tabID = tab.id;
            data.instance_id = this.tab.instance_id;
            data.connection_id = this.tab.connection_id;
        } else {
            this.logger.warn(`Tab ${tabID} not found`);
            data.tabID = tabID;
        }
        this.tabChannel.postMessage({ type, data, from: this.tab.id });
    }

    handleMessage(event) {
        const { type, data, from } = event.data;
        if (from === this.tab.id) {
            return; // Ignore self-sent messages
        }
        if (!this.tabs.has(from)) {
            this.tabs.set(from, new Tab(from, this.logger));
        }
        if (!this.tabs.has(data.tabID)) {
            this.tabs.set(data.tabID, new Tab(data.tabID, this.logger));
        }
        const dataTab = this.tabs.get(data.tabID);
        if (dataTab === this.masterTab && data.connection_id !== dataTab.connection_id) {
            this.logger.warn(`Master tab connection_id changed. Reconnecting.`);
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
                    this.electMaster();
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
                    this.logger.warn(`Removing inactive tab ${tabId}`);
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
        if (!this.tabs.has(this.tab.id)) {
            this.tabs.set(this.tab.id, this.tab);
        }
        const lowestId = Math.min(this.tab.id, ...Array.from(this.tabs.keys()).map(Number));
        if (lowestId === this.tab.id) {
            this.broadcast('master', this.tab.id);
            this.setMaster(this.tab.id);
        } else {
            this.setMaster(lowestId);
        }
    }

    isMaster() {
        return this.masterTab && this.masterTab.id === this.tab.id;
    }

    setMaster(masterID) {
        const wasMaster = this.isMaster();
        if (!this.masterTab || this.masterTab.id !== masterID) {
            this.masterTab = this.tabs.get(masterID);
            this.logger.info(`Master tab is now ${masterID}`);
        }
        const isNowMaster = this.isMaster();
        if (wasMaster !== isNowMaster) {
            this.updateStatus();
        }
    }

    updateStatus() {
        this.logger.info(`Status: ${this.isMaster() ? 'Master' : 'Slave'}`);
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
                    this.logger.error(`Error removing instance.`);
                    return;
                }
                this.logger.info(`Instance removed.`);
            };
            xhr.onerror = () => {
                this.logger.error('Failed to remove instance due to network error or server unavailability.');
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
            if (!this.masterTab) {
                setTimeout(() => {
                    this.logger.debug(`Waiting for master tab: ${JSON.stringify(this.masterTab)}`);
                    waitForMaster();
                }, 100);
                return;
            } else {
                this.logger.info(`Master tab found: ${JSON.stringify(this.masterTab)}`);
                this.setConnectionType();
                this.registerNewInstance();
            }
        };
        waitForMaster();
    }

    registerNewInstance() {
        const xhr = new XMLHttpRequest();
        if (this.masterTab.connection_id) {
            xhr.open(`GET`, `${this.url}/add/user?connection_id=${this.masterTab.connection_id}`);
        } else {
            xhr.open(`GET`, `${this.url}/add/user`);
        }
        xhr.send();
        xhr.onload = () => {
            if (xhr.status !== 200) {
                this.logger.error(`Error registering new instance.`);
                return;
            }
            const response = JSON.parse(xhr.responseText);
            this.tab.connection_id = response.connection_id;
            this.tab.instance_id = response.instance_id;
            this.startConnection();
        };
        xhr.onerror = () => {
            this.logger.error('Failed to register new instance due to network error or server unavailability.');
            if (this.onError) {
                this.onError();
            }
        };
    }

    changeConnection_id() {
        if (this.tab.instance_id && this.masterTab.connection_id) {
            this.setConnectionType();
            const xhr = new XMLHttpRequest();
            xhr.open(`GET`, `${this.url}/update/user?instance_id=${this.tab.instance_id}&connection_id=${this.masterTab.connection_id}`);
            xhr.send();
            xhr.onload = () => {
                if (xhr.status !== 200) {
                    this.logger.error(`Error changing connection_id.`);
                    this.open();
                    return;
                }
                this.logger.info(`Connection_id changed.`);
                this.tab.connection_id = this.masterTab.connection_id;
                this.startConnection();
            };
            xhr.onerror = () => {
                this.logger.error('Failed to change connection_id due to network error or server unavailability.');
                if (this.onError) {
                    this.onError();
                }
            };
        }
    }

    startConnection() {
        if (this.evtSource) {
            this.evtSource.close();
        }
        if (this.channel != null) {
            this.channel.close();
        }
        this.channel = new BroadcastChannel(`pubsubsse_communication`);
        this.channel.onmessage = null;
        this.status = `connected`;
        if (this.connection_type === `sse`) {
            this.connectWithSSE();
        } else if (this.connection_type === `broadcastchannel`) {
            if (this.masterTab.connection_id !== this.tab.connection_id) {
                this.changeConnection_id();
            }
            this.connectWithBroadcastChannel();
        }
    }

    connectWithSSE() {
        if (this.evtSource) {
            this.evtSource.close();
        }
        this.evtSource = new EventSource(`${this.url}/event?instance_id=${this.tab.instance_id}`);
        this.evtSource.onopen = () => {
            this.logger.info(`Connection to server opened.`);
            if (this.onConnected) {
                this.onConnected();
            }
            this.channel.postMessage({ type: `event_open` });
        };
        this.evtSource.onmessage = (e) => {
            const data = JSON.parse(e.data);
            this.logger.debug(`Received sse message: ` + e.data);
            this.channel.postMessage({ type: `data`, data });
            this.handleData(data);
        };
        this.evtSource.onerror = () => {
            this.logger.error(`EventSource failed.`);
            if (this.onError) {
                this.onError();
            }
            this.channel.postMessage({ type: `event_error` });
            this.open();
        };
        this.evtSource.onclose = () => {
            this.logger.info(`Connection to server closed.`);
            if (this.onDisconnected) {
                this.onDisconnected();
            }
            this.channel.postMessage({ type: `event_close` });
        };
    }

    connectWithBroadcastChannel() {
        this.channel.onmessage = (event) => {
            if (event.data.type === `data`) {
                const data = event.data.data;
                this.logger.debug(`Received broadcastchannel message: ` + JSON.stringify(data));
                this.handleData(data);
            }
            if (event.data.type === `event_open`) {
                this.logger.info(`Connection to server opened.`);
                if (this.onConnected) {
                    this.onConnected();
                }
            }
            if (event.data.type === `event_error`) {
                this.logger.error(`EventSource failed.`);
                if (this.onError) {
                    this.onError();
                }
            }
            if (event.data.type === `event_close`) {
                this.logger.info(`Connection to server closed.`);
                if (this.onDisconnected) {
                    this.onDisconnected();
                }
            }
        };
    }

    handleData(data) {
        const instances = data.instances;
        const instanceData = instances.find(instance => instance.id === this.tab.instance_id);
        if (!instanceData) {
            return;
        }
        if (this.isMaster()) {
            for (const instance of instances) {
                const instanceExists = Array.from(this.tabs.values()).some(tab => tab.instance_id === instance.id);
                if (!instanceExists) {
                    this.logger.warn(`Instance not found: ` + instance.id);
                    const attempts = this.unknownInstances.get(instance.id) || 0;
                    this.unknownInstances.set(instance.id, attempts + 1);
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
            if (type === `topics`) {
                let removedTopicsList = sysData.list;
                sysData.list.forEach(topicInfo => {
                    this.ensureTopic(topicInfo.id, topicInfo.type);
                    delete removedTopicsList[topicInfo.id];
                });
                removedTopicsList.forEach(topicId => {
                    const topic = this.topics.get(topicId);
                    if (topic) {
                        if (topic.subscribed) {
                            topic.onUnsubscribed?.();
                        }
                        this.topics.delete(topicId);
                        this.onRemovedTopic?.(topic);
                    }
                });
            } else if (type === `subscribed`) {
                sysData.list.forEach(topicInfo => {
                    const topic = this.ensureTopic(topicInfo.ID, topicInfo.type);
                    topic.onSubscribed?.();
                    topic.subscribed = true;
                });
            } else if (type === `unsubscribed`) {
                sysData.list.forEach(topicInfo => {
                    const topic = this.topics.get(topicInfo.ID);
                    if (topic) {
                        topic.onUnsubscribed?.();
                        topic.subscribed = false;
                    }
                });
            }
        });
    }

    handleUpdateMessages(updateData) {
        if (!updateData) return;
        updateData.forEach(update => {
            const topic = this.topics.get(update.topic);
            if (topic) {
                topic.onUpdate?.(update.data);
            }
        });
    }

    ensureTopic(ID, type) {
        let topic;
        if (!this.topics.has(ID)) {
            this.logger.debug(`New topic: ` + ID);
            topic = new Topic(ID, type, this.logger);
            this.topics.set(ID, topic);
            if (this.onNewTopic) {
                this.onNewTopic?.(topic);
            }
        } else {
            topic = this.topics.get(ID);
        }
        return topic;
    }
}
