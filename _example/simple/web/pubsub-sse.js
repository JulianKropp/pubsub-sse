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

class PubSubSSE {
    constructor(url = `http://localhost`) {
        this.url = url;
        this.connection_type = `sse`; // sse or broadcastchannel
        this.channel = new BroadcastChannel(`pubsub-sse-channel`);
        this.instance_id = null;
        this.connection_id = null;
        this.connection_id_by_broadcastchannel = null;
        this.instance_id_by_broadcastchannel = null;
        this.evtSource = null;
        this.topics = {}; // Stores Topic objects

        this.onConnected = null;
        this.onDisconnected = null;
        this.onError = null;
        this.onNewTopic = null;
        this.onRemovedTopic = null;
    }

    open() {
        this.instance_id = null;
        this.connection_id = null;
        this.connection_id_by_broadcastchannel = null;
        this.instance_id_by_broadcastchannel = null;

        // Send a message to the channel asking for other tabs
        this.channel.postMessage({ type: `request_presence` });

        console.log(`Requesting presence. Waiting for response.`);

        // Listen for responses or lack thereof
        const responseTimeout = setTimeout(() => {
            if (!this.connection_id) { // If no response in 100ms, make a new user request

                console.log(`No response received. Registering new instance.`);

                this.connection_type = `sse`;
                this.registerNewInstance();
            }
        }, 100);

        this.channel.onmessageerror = (event) => {
            console.log(`Error with broadcastchannel: ` + event);
        };

        this.channel.onmessage = (event) => {
            if (event.data.type === `response_presence` && !this.connection_id) {

                console.log(`Received response broadcastchannel message: ` + JSON.stringify(event.data));

                // Cancel the timeout as we have a response
                clearTimeout(responseTimeout);
                this.connection_type = `broadcastchannel`;
                this.connection_id_by_broadcastchannel = event.data.connection_id;
                this.instance_id_by_broadcastchannel = event.data.instance_id;

                this.registerNewInstance();
            }
        };
    }

    registerNewInstance() {
        const xhr = new XMLHttpRequest();
        if (this.connection_id_by_broadcastchannel) {
            // This will register a new instance and will send all the data to the connection_id which is normaly anouther tab (instance)
            // If the connection_id is not valid, the server will return a new connection_id
            xhr.open(`GET`, `${this.url}/add/user?connection_id=${this.connection_id_by_broadcastchannel}`);
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
            this.connection_id = response.connection_id;
            this.instance_id = response.instance_id;
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

        // Open new connection based on connection type
        if (this.connection_type === `sse`) {
            this.connectWithSSE(); // Open a new SSE connection
        } else if (this.connection_type === `broadcastchannel`) {
            if (this.connection_id_by_broadcastchannel === this.connection_id) {
                this.connectWithBroadcastChannel(); // Listen for messages on the BroadcastChannel
            } else {
                // Fall back to SSE if the connection_id is not valid
                this.connectWithSSE();
            }
        }

        // Once everything is set up, prepare to handle future requests
        if (this.connection_type === `sse`) {
            this.channel.onmessage = (event) => {
                if (event.data.type === 'request_presence') {
                    this.channel.postMessage({
                        type: 'response_presence',
                        connection_id: this.connection_id,
                        instance_id: this.instance_id
                    });
                }
            };
        }
    }

    // Open a new SSE connection
    connectWithSSE() {
        // close existing connection
        if (this.evtSource) {
            this.evtSource.close();
        }

        this.evtSource = new EventSource(`${this.url}/event?instance_id=${this.instance_id}`);

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
        const instanceData = data.instances.find(instance => instance.id === this.instance_id);
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