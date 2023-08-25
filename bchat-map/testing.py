import paho.mqtt.client as mqtt
from decimal import Decimal
import datetime

def on_connect(client, userdata, flags, rc):
    print("Yay test connected")

mqtt_host="localhost"
mqtt_port=1883
mqtt_id="testing"

client = mqtt.Client(client_id = mqtt_id)
client.on_connect = on_connect

# Blocking connect call
print("Attempting to connect to {mqtt_host}:{mqtt_port} using id '{mqtt_id}'...".format(mqtt_host=mqtt_host, mqtt_port=mqtt_port, mqtt_id=mqtt_id))
client.connect(host=mqtt_host, port=mqtt_port)
client.loop_start()

print("Started loop")
GPS_TOPIC = "bchat/rooms/main/sheep_loc"
ll_temple = (Decimal('40.791255'), Decimal('-119.197142'))
lat_long_message = "{lat:3.7f},{long:3.7f}"
client.publish(GPS_TOPIC, lat_long_message.format(lat=ll_temple[0], long=ll_temple[1]))

class Message:
    def __init__(self):
        self.Sent = datetime.datetime.now()
        self.From = "testing"
        self.Msg = "this is a test"
        
MESSAGES_TOPIC = "bchat/rooms/main/testing/messages"
# client.publish(MESSAGES_TOPIC, Message())

print("Published")