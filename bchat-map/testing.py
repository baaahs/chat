import paho.mqtt.client as mqtt
from decimal import Decimal
import datetime
import time
import json

def on_connect(client, userdata, flags, rc):
    print("Yay test connected")

    GPS_TOPIC = "bchat/rooms/main/sheep_loc"
    ll_temple = (Decimal('40.791255'), Decimal('-119.197142'))
    lat_long_message = "{lat:.7f},{long:.7f}"

    msg_str = lat_long_message.format(lat=ll_temple[0], long=ll_temple[1])
    print("Publishing ", msg_str)
    client.publish(GPS_TOPIC, msg_str)
    print("That is queued")


def on_publish(a,b,c):
    print("on_publish called, Publish is done")

mqtt_host="tompop.tomseago.com"
mqtt_port=1883
mqtt_id="testing"

client = mqtt.Client(client_id = mqtt_id)
client.on_connect = on_connect
client.on_publish = on_publish

# Blocking connect call
print("Attempting to connect to {mqtt_host}:{mqtt_port} using id '{mqtt_id}'...".format(mqtt_host=mqtt_host, mqtt_port=mqtt_port, mqtt_id=mqtt_id))
client.connect(host=mqtt_host, port=mqtt_port)
client.loop_start()

print("Started loop")
message = {
    "msg": "this is a test message",
    "sent": time.time(),
    "from": "testing" 
}
        
MESSAGES_TOPIC = "bchat/rooms/main/*/messages"
client.publish(MESSAGES_TOPIC, json.dumps(message))

message = {
    "msg": "this is a large test message that is larger than the max characters per line message",
    "sent": time.time(),
    "from": "testing" 
}
client.publish(MESSAGES_TOPIC, json.dumps(message))

print("Published")

while True:
    time.sleep(1)