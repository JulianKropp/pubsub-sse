```uml
@startuml
package pubsub-sse {

class client {
-id: string
-status: status
-stream: chan string
-onEvent: OnEventFunc
-stopchan: chan interface{}
-lock: Mutex
-sSEPubSubService: *sSEPubSubService
-privateTopics: map[string]*topic
-groupTopics: map[string]*topic
-stop()
+GetID(): string
+GetStatus(): status
+GetPublicTopics(): map[string]*topic
+GetPublicTopicByName(name string): *topic, bool
+GetPrivateTopics(): map[string]*topic
+GetPrivateTopicByName(name string): *topic, bool
+GetGroups(): map[string]*topic
+GetAllTopics(): map[string]*topic
+GetTopicByName(name string): *topic, bool
+GetSubscribedTopics(): map[string]*topic
+NewPrivateTopic(name string): *topic
+RemovePrivateTopic(t *topic)
+Sub(topic *topic): error
+Unsub(topic *topic): error
-send(msg interface): error
-sendTopicList(): error
-sendSubscribedTopic(topic *topic): error
-sendUnsubscribedTopic(topic *topic): error
-sendInitMSG(): error
+OnEvent(f OnEventFunc)
+RemoveOnEvent()
+Start(ctx Context)
}
class status {


}
class OnEventFunc {


}
class sSEPubSubService {
-clients: map[string]*client
-publicTopics: map[string]*topic
-lock: Mutex
+NewClient(): *client
+RemoveClient(c *client)
+GetClients(): map[string]*client
+GetClientByID(id string): *client, bool
+NewPublicTopic(name string): *topic
+RemovePublicTopic(t *topic)
+GetPublicTopics(): map[string]*topic
+GetPublicTopicByName(name string): *topic, bool
}
class topic {
-name: string
-id: string
-ttype: topicType
-clients: map[string]*client
-lock: Mutex
+GetName(): string
+GetID(): string
+GetType(): string
-addClient(c *client)
-removeClient(c *client)
+GetClients(): map[string]*client
+IsSubscribed(c *client): bool
+Pub(msg interface): error
}
class eventDataSys {
+Type: string
+List: []eventDataSysList

}
class eventDataSysList {
+Name: string
+Type: string

}
class topicType {


}
class eventData {
+Sys: []eventDataSys
+Updates: []eventDataUpdates

}
class eventDataUpdates {
+Topic: string
+Data: interface

}
}
"client" --> "status"
"client" --> "OnEventFunc"
"client" --> "sSEPubSubService"
"client" --> "topic"
"sSEPubSubService" --> "client"
"sSEPubSubService" --> "topic"
"topic" --> "topicType"
"topic" --> "client"
"eventDataSys" --> "eventDataSysList"
"eventDataUpdates" --> "topic"
"eventData" --> "eventDataSys"
"eventData" --> "eventDataUpdates"

@enduml
```