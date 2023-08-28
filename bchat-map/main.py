import tkinter as tk
from tkinter import PhotoImage
from decimal import Decimal
import os
import paho.mqtt.client as mqtt
import datetime
import json

# Parse command line args for fun and profit (and to make development life easier)
import argparse

parser = argparse.ArgumentParser(description="BAAAHS Chat map display")
parser.add_argument("--dev", action="store_true", help="Run in development mode (not fullscreen)")
parser.add_argument("--add_test_points", action="store_true", help="Add a bunch of test points to the map")
parser.add_argument("--mqtt_host", default="localhost", type=str, help="MQTT host address")
parser.add_argument("--mqtt_port", default=1883, type=int, help="MQTT port number")
parser.add_argument("--mqtt_id", default="map_test", type=str, help="MQTT client ID")
args = parser.parse_args()


# MQTT stuff
def on_connect(client, userdata, flags, rc):
    print("Yay map connected")

client = mqtt.Client(client_id = args.mqtt_id)
client.on_connect = on_connect

# Blocking connect call
print("Attempting to connect to {mqtt_host}:{mqtt_port} using id '{mqtt_id}'...".format(mqtt_host=args.mqtt_host, mqtt_port=args.mqtt_port, mqtt_id=args.mqtt_id))
client.connect(host=args.mqtt_host, port=args.mqtt_port)
client.loop_start()

# Coordinates for the map_1080p.png file. Tweaks have been made visually
# using the test points to get things to mostly line up
lat_max_bound = lat_upper_left = Decimal('40.804187')
long_min_bound = long_upper_left = Decimal('-119.232332')
lat_min_bound = lat_bottom_right = Decimal('40.769762')
long_max_bound = long_bottom_right = Decimal('-119.152574')

lat_max_y = lat_max_bound - lat_min_bound
long_max_x = long_max_bound - long_min_bound

# From the 2023 KMZ dataset directly
ll_man = (Decimal('40.786393'), Decimal('-119.203515'))
temple = ll_temple = (Decimal('40.791255'), Decimal('-119.197142'))

# 4:00 and C 
ll_four_and_C = (Decimal("40.777589"), Decimal("-119.200443"))

# Consts
max_message_chars = 40
max_message_list_size = 25
new_gps_event_name = "<<NEW-GPS-COORDS>>"
new_gps_message_format = "Last BAAAHS Location @ {when}"

def split_message(message: str):
    message_len = len(message)
    if (message_len <= max_message_chars):
        return [message]

    split_msg = []
    words = message.split(" ")
    current_line = ""

    for idx, w in enumerate(words):
        current_line += w + " "
        if (len(current_line) > max_message_chars):
            split_msg.append(current_line.rstrip())
            current_line = "\t"

        if (idx == len(words)-1):
            split_msg.append(current_line.rstrip())

    return split_msg

class BaaahsMap:
    
    def __init__(self, image, icon, man, crosshair):
        self.map_file = image
        self.icon_file = icon
        self.man_file = man
        self.crosshair_file = crosshair
        self.root = tk.Tk()

        # Only do fullscreen if not in "dev" mode
        if not args.dev:
            self.root.attributes("-fullscreen", True)
        self.root.title("Shee-PS")

        self.baaahs_pos_ll = ll_four_and_C

        self.widgets()

    def mainLoop(self):
        self.root.mainloop()
        
    def widgets(self):
        # Map Stuff
        self.map_img = tk.PhotoImage(file=self.map_file)
        self.bgMap_width = Decimal(self.map_img.width())
        self.bgMap_height = Decimal(self.map_img.height())
        print("bgMap_width=",self.bgMap_width,"  bgMap_height=",self.bgMap_height)

        self.baaahs = tk.PhotoImage(file=self.icon_file, width=40, height=34)
        # self.man_img = tk.PhotoImage(file=self.man_file, width=45, height=45)
        self.man_img = tk.PhotoImage(file=self.man_file)
        self.crosshair_img = tk.PhotoImage(file=self.crosshair_file)
        
        self.canvas = tk.Canvas(self.root, width=self.map_img.width(), height=self.map_img.height())
        self.canvas.create_image(0, 0, anchor='nw', image=self.map_img)

        # NOTE ABOUT ANCHORS
        # Originally the normalize_coords function would try to normalize to a corner
        # of a particular image. But this isn't necessary if we just use the tk.CENTER
        # anchor constant for all the images

        # The Man
        man_pos = self.normalize_coords(ll_man[0], ll_man[1])
        self.canvas.create_image(man_pos[0], man_pos[1], anchor=tk.CENTER, image=self.man_img)

        # Initial BAAAHS
        self.baaahs_pos = self.normalize_coords(self.baaahs_pos_ll[0], self.baaahs_pos_ll[1])
        self.baaahs_id = self.canvas.create_image(self.baaahs_pos[0], self.baaahs_pos[1], anchor=tk.CENTER, image=self.baaahs)

        self.canvas.event_add(new_gps_event_name, "None")
        self.canvas.bind(new_gps_event_name, self.new_baaahs_coords)

        # Add test point crosshairs if desired
        if args.add_test_points:
            self.add_test_points()

        # Right Text List
        text_box_width = self.map_img.width() / 6
        text_box_height = self.map_img.height() / 2
        self.text_box = tk.Listbox(self.root, bg='#000', font='arial', fg="#fff")
        self.text_box.pack()

        text_box_pos_x = self.map_img.width() # text_box_width
        text_box_pos_y = 0
        self.canvas.create_window(text_box_pos_x, text_box_pos_y, anchor=tk.NE, window=self.text_box, width=text_box_width, height=text_box_height)
        #self.canvas.focus()

        # Show text box over map
        self.text_box.focus()

        # Finally pack
        self.canvas.pack()

    def set_new_baaahs_cords(self, lat, long):
        print("new baahs coords", lat, long)
        self.baaahs_pos_ll = (Decimal(lat), Decimal(long))
        self.canvas.event_generate(new_gps_event_name)

    def new_baaahs_coords(self, event):
        self.baaahs_pos = self.normalize_coords(self.baaahs_pos_ll[0], self.baaahs_pos_ll[1])
        self.canvas.delete(self.baaahs_id)
        self.baaahs_id = self.canvas.create_image(self.baaahs_pos[0], self.baaahs_pos[1], anchor=tk.CENTER, image=self.baaahs)
        self.add_message_to_box(new_gps_message_format.format(when=datetime.datetime.now(), lat=self.baaahs_pos_ll[0], long=self.baaahs_pos_ll[1]))
    
    def add_message_to_box(self, message):
        print("new baahs message", message)
        split_msg = split_message(message)
        for msg in split_msg:
            self.text_box.insert(tk.END, msg)
            if self.text_box.size() > max_message_list_size:
                self.text_box.delete(0)

        
    
    def normalize_coords(self, lat, long):
        norm_lat = (lat - lat_min_bound)/lat_max_y
        norm_long = (long - long_min_bound)/long_max_x
        img_x = norm_long * self.bgMap_width
        img_y = norm_lat * self.bgMap_height

        # Verbose, but consistent
        output_x = img_x
        output_y = img_y

        # Invert the Y coordinate to make up for the differences in counting from 
        # "bottom to top" as latitude does versus "top to bottom" as most common
        # output displays do
        output_y = self.bgMap_height - output_y

        return (output_x, output_y)


    def add_test_point(self, lat, long):
        point_loc = self.normalize_coords(lat, long)
        self.canvas.create_image(point_loc[0], point_loc[1], anchor=tk.CENTER, image=self.crosshair_img)

    def add_test_points(self):
        # United Site Services
        self.add_test_point(Decimal("40.777"),Decimal("-119.223849"))
        # Greeters
        self.add_test_point(Decimal("40.773028"),Decimal("-119.220986"))
        # 4:30 & G Plaza
        self.add_test_point(Decimal("40.772994"),Decimal("-119.203467"))
        # Point 4
        self.add_test_point(Decimal("40.776026"),Decimal("-119.17628"))
        # Point 3
        self.add_test_point(Decimal("40.80288"),Decimal("-119.182115"))
        # Hell Station
        self.add_test_point(Decimal("40.803056"),Decimal("-119.209183"))
        # 1200 Promenade
        self.add_test_point(Decimal("40.788818"),Decimal("-119.200315"))
        # 730 Portal
        self.add_test_point(Decimal("40.786374"),Decimal("-119.212529"))
        # 430 Portal
        self.add_test_point(Decimal("40.779526"),Decimal("-119.203484"))

        # And why not these as well???
        self.add_test_point(ll_man[0], ll_man[1])
        self.add_test_point(ll_temple[0], ll_temple[1])

        self.add_test_point(lat_min_bound + ((lat_max_bound - lat_min_bound)/2),
                            long_min_bound + ((long_max_bound - long_min_bound)/2))

MESSAGES_TOPIC = "bchat/rooms/main/*/messages"
GPS_TOPIC = "bchat/rooms/main/sheep_loc"

dir = os.getcwd()
#image = dir + "/bchat-map/resources/brc_map_better.pgm"
image = dir + "/bchat-map/resources/map_1080p.png"
icon = dir + "/bchat-map/resources/baaahs_icon.pgm"
man = dir + "/bchat-map/resources/the_man.pgm"
crosshair = dir + "/bchat-map/resources/crosshair.png"

baaahsMap = BaaahsMap(image, icon, man, crosshair)

def on_new_sheep_coords(client, userdata, message):
    new_ll = message.payload.decode().split(",")
    new_lat = new_ll[0]
    new_long = new_ll[1]
    baaahsMap.set_new_baaahs_cords(new_lat, new_long)

message_format = "[{date}]: [{who}]: {text}"
def on_new_message(client, userdata, message):
    data = json.loads(message.payload.decode())
    date = datetime.datetime.fromtimestamp(int(data["sent"]))
    baaahsMap.add_message_to_box(message_format.format(date=date, who=data["from"], text=data["msg"]))

# Subs to GPS
client.subscribe(GPS_TOPIC, qos=0)
client.message_callback_add(GPS_TOPIC, on_new_sheep_coords)

# Subs to Messages
client.subscribe(MESSAGES_TOPIC, qos=0)
client.message_callback_add(MESSAGES_TOPIC, on_new_message)

baaahsMap.mainLoop()