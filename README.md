BAAAHS Chat (B-Chat)
====================

This repo defines the chat server, protocol, and some (maybe all) of the clients for the BAAAHS chat system between the Sheep and Camp. The primary "art" that is being built around this for 2019 is to enable two low end chromebooks, one on the sheep and one in camp, to be able to send messages back and forth. We also expect to be able to have other clients connect such as a small RPI with a GPS that can tell everyone where it thinks it is at on a periodic basis.

This system uses some of the same network as Sparkle Motion but is being built separately from the lighting project and does not rely on it. There will almost certainly be connection between the two over time.


## Theory of Operation

![](doc/basic-network.png)

At the core is the network. The network is conceptually two separate partions - one on the sheep and one at camp - that periodically manage to talk to each other. Hopefully all the time, but it's quite possible that this link will be intermittent. 

Additionally, from a user perspective this chat implementation is intentionally going to be very constrained. Basically it will be a one room IRC like service. We decided not to use regular IRC because it already has to many features and we don't want the distraction.

For transport reliability we're using [MQTT] servers, one on the sheep and one at the camp, with a bridge between them for exchanging messages between all the user-agent clients. We are using what is essentially the reference [MQTT] server named [mosquitto].

## Development Setup

You intentionally need very little to get started. You could use the spec defined below and none of the code from this repo, but you can also use the code from this repo.

You're going to need an MQTT broker to talk to, so you'll want to install [mosquitto]. On a reasonably modern Mac with [brew] installed you should be able to install [mosquitto] by

    brew install mosquitto
    
On my main machine the brew installation complained about a missing `/usr/local/sbin` that I had to go create by hand and then do a `brew link mosquitto` after I had done that. Hopefully this is a local me issue though.

With this repo checked out and [mosquitto] installed, from the `mosquitto` directory of this repo you should be able to start a local [mosquitto] server that binds to all interfaces of your machine by running

    /usr/local/sbin/mosquitto -c local.conf
    
That should give you a totally default server running on port `1883`. The `local2.conf` file can be used to start a second instance on a different port of `18830` which will then bridge itself to the first server and will pass all messages between them.

Now you need a client. The `bchat-tty` directory contains the code for a client written in go that we are using for the Chromebooks. You can also run this client locally multiple times in multiple terminal windows to get the real effect. 

If you just want to run a prebuilt version of the client try the version in the `bin` directory

    bin/bchat-tty-osx
    
That will _probably_ run on any Mojave machine. If it doesn't then you'll have to setup a dev environment for go, which isn't hard, but is beyond the current scope of this readme. So let's say that worked and you have a chat client running.

If you are running an [MQTT] server on `localhost:1883` then it should have connected to it and you will see any messages sent on the topics described below. If you don't have a local server running, start one, and it should connect.

***

When clients connect to an [MQTT] server they use a client id. [bchat-tty] uses the local machine id as it's client id which is good for the chromebooks but maybe not what you want in development because you want to run multiple copies of it in multiple windows. This id is separate from the "nick" that it uses when sending messages. The idea being that the "nick" might be user modifiable but this machine id is unique to the machine. This let's the [MQTT] server know which messages have gone to which machines and do all the right things to get the messages delivered.

So in order to multiple copies of [bchat-tty] you need to provide it some configuration which is done via the [Archer Configuration Language][archercl]. ACL is basically "json plus comments" and can be provided from files, from the command line, or from the environment. Instead of command line arguments, `bchat-tty` just uses the ACL infrastructure to figure out how it should behave.

> The main thing you need to know about ACL is that keys and values can be separated by colons or equals signs, and each command line argument is parsed as an individual string, so for simple non-whitespace keys and values you can set configuration keys from the command line like so.
>
>     key1:value1 key2=value2
>
> And if you needed more complex values you might write something like the following as a command line argument. The single quotes protect the string from the shell and the double quotes protect the url with it's colons from the ACL parser. 
>
>     'url: "tcp://10.0.1.1:1883"'

Thus, to run multiple clients on one machine, all you need is to to provide the different processes with slightly different configurations. So in one terminal window you might run

    bchat-tty-osx id:1 name:Sheep
    
And in another terminal you could run 

    bchat-tty-osx id:2 name:Camp
    
That should get you two terminals that can talk to each other.

> Note that ***all clients must have unique ids***. If an [MQTT] broker receives two TCP connections that both present the same client id, it assumes they are the same client and will drop the old connection. Since [bchat-tty] will auto reconnect if you run it twice without providing separate ids, the two instances will keep kicking each other off and things won't work.

With two clients running, and chatting back and forth, you should be able to add additional clients into the system that publish messages on topics as described below and they should show up in the running clients.

If you want to tell [bchat-tty] to connect to a broker somewhere other than localhost, you need to change the value of the **url** key.The **url** has to be a go formatted network address and if you're doing it from the command line pay attention to the escaping requirements as shown above.

If you want to specify configuration values using the environment, you would do that like so:

    export bchat_id=1000
    export bchat_name=Fred
    export bchat_url="tcp://localhost:18830"
    
And if you want to use configuration files you need to know that the ACL infrastructure loads a cascade of files starting with a system level, then a user specific file, then finally a directory specific file. After that it loads any files that are specified on the command line using the `-c` or `--config` command line parameters. 

    bchat-tty-mac -c bob.acl 
    
Will load the following files if they exist in this order:

    /etc/bchat.acl
    ~/.bchat.acl
    ./bchat.acl
    bob.acl

As mentioned earlier, ACL is essentially JSON-ish but allows comments in pretty much all styles, so if you wanted to setup some config files for your different instances an almost complete example of everything that [bchat-tty] understands with the default values it uses would be

    // The url for the MQTT server. Must be specified as a valid go tcp:// url.
    // Note that in ACL values that contain :'s must be quoted.
    url: "tcp://localhost:1883" 
    
    // The id to use. If not specified in a config file will default to a
    // unique machine identifier. How this is created is based on the hardware
    // and operating system of the machine.
    id: <default depends on the machine>
    
    // Default nick name to use until the user changes it
    name: Sheep
    
    // Whether to show the in-app log window by default
    log: false
    
The ACL code also recognizes a standard set of logging configuration values. See the [acl module documentation][archercl] if you're curious about those.

If you want to monitor what's going on in terms of messages being passed etc. there are various [MQTT tools](https://github.com/mqtt/mqtt.github.io/wiki/tools) available that might help you out. Since there are [MQTT client libraries](https://github.com/mqtt/mqtt.github.io/wiki/libraries) for pretty much every language this infrastructure can probably be used for all sorts of other projects as well.


## Topics


[MQTT] is a pub/sub system with a hiearchy of topics to which arbitrary messages may be published. Subscribers can subscribe using wildcards to get various parts of a potentially dynamic tree.

We're going to define the tree with more functionality than we plan to implement in rev 1 because the future always comes faster than you think. So while we are planning to only expose one room, we'll make it easy to add more later. Thus, a chat room has a place in the topic tree that starts with `bchat/rooms/`. The next segment in the topic is the room name. For rev 1, the only room anyone should use is **main** so the full topic for the room is `bchat/rooms/main`. 

Underneath each room we will have sub-topics as follows:

   - **.../{nick}/messages** - Messages sent by a given nick (Sheep or Station etc)
     
     It's likely that the user will be allowed to change the nick so this may or may not be a good idea. We maybe should use the client id instead. We'll figure it out later but presume there is one level of hierarchy here.
     
   - **.../{nick}/status** - Per user status information
   
     I don't know all of what will go into this message but this gives us a spot to dump a retained message indicating what was last known from this nick / location. Could be interesting technical data but could also contain things like what song is playing maybe? Although that specific data would be better served on it's own topic I think. 


In the future we might add things like a list of users or stats or something. Again - keep it simple for now. Notice that we are including the client name in the topic. I saw this in a best practice recommendation and it seems kind of reasonable. Because a client can use wildcards in subscriptions it seems like it makes sense. 

A basic chat client will subscribe to `bchat/rooms/main/+/messages` and then just show everything it receives. Similarly it will publish it's own messages to `bchat/rooms/main/{nick}/messages` where `{nick}` might be something like `Sheep` 

When it comes to payloads, [MQTT] is agnostic, but let's just go ahead and say that we prefer payloads to be JSON. Special cases can use other formats if they want, but if JSON can do it, then it should be used.

We may want to post position updates both to the chat as messages as well as to another topic in more machine readable form. If we do that, let's declare that the topic will be `sheep/position/updates` with a JSON format that can be defined when we actually do it. These messages can be sent as retained messages and then the magic of [MQTT] should always get the last sent one to a intermittent client (like one running on a cell phone).

It's also reasonable that Pinky might want to publish song meta data that it gets from the CDJs to `sheep/tracks/current`. If it does then any client that wants to display the info could subscribe.

We probably want some status information from the clients and each of the brokers. Need to figure out a reasonable topic tree for that. I'm thinking something where they can send retained LWT messages so that when connections drop we can show that in a UI of a client. We'll see about this. It's not hard, but it's also not all that important.


### Message Payloads

Only this for now. These are published to `bchat/rooms/{room name}/{nick}/messages` 

    {
        "msg": string - Contents of the message. Any UTF string
        "sent": integer - Unix timestamp for when the message was sent 
                          (seconds since Jan 1, 1970)
        "from": string - Nickname/Id of the sender. Will be displayed as the source
                         of the message.
    }

### Behaviors


We're keeping this super trusty and simple to begin with. No authentication of clients is required, and one client could totally impersonate another because the from of each message is generated by the sender. Clients should behave unless it's funny for them not to. This is Burning Man after all.

It would be nice if clients published some basic status info like "connected" or something like that. They could also send will messages so we know when they go away.

### QOS

I'm thinking that the whole point of [MQTT] is the store and forward capabilities that we get at QOS level 1 and 2. The only reason not to do 2 is "it can take a little longer", but hey, at our scale I don't think that's an issue. So let's send everything at QOS level 2 baby! (That's the highest level)

Most clients should also **not** set the "clean session" flag when establishing the connection to the [MQTT] broker. This means the broker will store any messages sent to topics the client was subscribed to when it goes away and will send them to the client when it comes back. If the client is able to know that it's not likely to come back, it should definitely unsubscribe from all it's topics so the server doesn't bother with this. It's not a big deal if it doesn't because we'll configure something reasonable on the server to drop messages after like a day, but it we might as well be nice.
 
## bchat-tty

Think of this as the reference implementation of the bchat system.

We'll write more about it later.

[mosquitto]: http://mosquitto.org
[brew]: https://brew.sh
[archercl]: https://github.com/eyethereal/go-archercl
[bchat-tty]: #bchat-tty
[mqtt]: http://mqtt.org