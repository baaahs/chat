BAAAHS Chat (B-Chat)
====================

This repo defines the chat server, protocol, and some (maybe all) of the clients for the BAAAHS chat system between the Sheep and Camp. The primary "art" that is being built around this for 2019 is to enable two low end chromebooks, one on the sheep and one in camp, to be able to send messages back and forth. We also expect to be able to have other clients connect such as a small RPI with a GPS that can tell everyone where it thinks it is at on a periodic basis.

This system uses some of the same network as Sparkle Motion but is being built separately from the lighting project and does not rely on it. There will almost certainly be connection between the two over time.


## Theory of Operation


At the core is the network. The network is conceptually two separate partions - one on the sheep and one at camp - that periodically manage to talk to each other. Hopefully all the time, but it's quite possible that this link will be intermittent. 

Additionally, from a user perspective this chat implementation is intentionally going to be very constrained. Basically it will be a one room IRC like service. We decided not to use regular IRC because it already has to many features and we don't want the distraction.

For transport reliability we're going to try to use MQTT servers, one on the sheep and one at the camp, with a bridge between them for exchanging messages between all the user-agent clients.


## Topics


MQTT is a pub/sub system with a hiearchy of topics to which arbitrary messages may be published. Subscribers can subscribe using wildcards to get various parts of a potentially dynamic tree.

We're going to define the tree with more functionality than we plan to implement in rev 1 because the future always comes faster than you think. So while we are planning to only expose one room, we'll make it easy to add more later. Thus, a chat room has a place in the topic tree that starts with `bchat/rooms/`. The next segment in the topic is the room name. For rev 1, the only room anyone should use is `bchat/rooms/main`. 

Underneath each room we will have sub-topics as follows:

- **.../{nick}/messages** - Messages sent by a given nick (Sheep or Station etc)
- **.../{nick}/status** - I don't know all of what will go into this message but this gives us a spot to dump a retained message indicating what was last known from this nick / location. Could be interesting technical data but could also contain things like what song is playing 

In the future we might add things like a list of users or stats or something. Again - keep it simple for now. Notice that we are including the client name in the topic. I saw this in a best practice recommendation and it seems kind of reasonable. Because a client can use wildcards in subscriptions it seems like it makes sense. 

For instance, to do basic chat a client can subscribe to `bchat/rooms/main/+/messages` Yeah, this feels right. Using wildcard matching where it already exists in other services.

When it comes to payloads, MQTT is agnostic, but let's just go ahead and say that we prefer payloads to be JSON.

We may want to post position updates both to the chat as messages as well as to another topic in more machine readable form. If we do that, let's declare that the topic will be `sheep/position/updates` with a JSON format that can be defined when we actually do it. These messages can be sent as retained messages and then the magic of MQTT should always get the last sent one to a intermittent client (like one running on a cell phone).

We probably want some status information from the clients and each of the brokers. Need to figure out a reasonable topic tree for that. I'm thinking something where they can send retained LWT messages so that when connections drop we can show that in a UI of a client. We'll see about this. It's not hard, but it's also not all that important.


### Message Payloads

Only this for now. 

    {
        "msg": string - Contents of the message
        "sent": integer - Unix timestamp for when the message was sent 
                          (seconds since Jan 1, 1970)
        "from": string - Nickname/Id of the sender. Will be displayed as the source
                         of the message.
    }

Behaviors


We're keeping this super trusty and simple to begin with. No authentication of clients is required, and one client could totally impersonate another because the from of each message is generated by the sender. Clients should behave unless it's funny for them not to. This is Burning Man after all.

It would be nice if clients published some basic status info like "connected" or something like that. They could also send will messages so we know when they go away.

### QOS

I'm thinking that the whole point of MQTT is the store and forward capabilities that we get at QOS level 1 and 2. The only reason not to do 2 is "it can take a little longer", but hey, at our scale I don't think that's an issue. So let's send everything at QOS level 2 baby! (That's the highest level)

