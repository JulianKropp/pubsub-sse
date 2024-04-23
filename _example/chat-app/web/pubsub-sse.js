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
    constructor(url) {
        this.url = url;
        this.evtSource = null;
        this.topics = {}; // Stores Topic objects

        this.onConnected = null;
        this.onDisconnected = null;
        this.onError = null;
        this.onNewTopic = null;
        this.onRemovedTopic = null;
    }

    open() {
        if (this.evtSource) {
            this.evtSource.close();
        }

        this.evtSource = new EventSource(this.url);

        this.evtSource.onopen = () => {
            console.log("Connection to server opened.");
            if (this.onConnected) {
                this.onConnected();
            }
        };

        this.evtSource.onmessage = (e) => {
            const data = JSON.parse(e.data);

            console.log("Received message: " + e.data);

            this.handleSysMessages(data.sys);
            this.handleUpdateMessages(data.updates);
        };

        this.evtSource.onerror = () => {
            console.log("EventSource failed.");
            if (this.onError) {
                this.onError();
            }
        };

        this.evtSource.onclose = () => {
            console.log("Connection to server closed.");
            if (this.onDisconnected) {
                this.onDisconnected();
            }
        }
    }

    handleSysMessages(sysDatas) {
        if (!sysDatas) return;

        sysDatas.forEach(sysData => {
            const type = sysData.type;
            // Types:
            //   "topics": List of topics
            //   "subscribed": List of subscribed topics
            //   "unsubscribed": List of unsubscribed topics

            if (type === "topics") {
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
            } else if (type === "subscribed") {
                sysData.list.forEach(topicInfo => {
                    const topic = this.ensureTopic(topicInfo.ID, topicInfo.type);
                    topic.onSubscribed?.(); // Call the onSubscribed event if defined
                    topic.subscribed = true; // Mark as subscribed
                });
            } else if (type === "unsubscribed") {
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
            console.log("New topic: " + ID);
            topic = new Topic(ID, type); // Assign the new Topic to the local variable
            this.topics[ID] = topic; // Store the Topic in the topics dictionary
            this.onNewTopic?.(topic); // Notify client of new topic using the local variable
        } else {
            topic = this.topics[ID]; // If the topic already exists, assign it to the local variable
        }
        return topic;
    }
    
}