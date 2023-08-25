import tkinter as tk
from tkinter import PhotoImage
from decimal import Decimal
import os
import paho.mqtt.client as mqtt
import datetime

# MQTT stuff
def on_connect(client, userdata, flags, rc):
    print("Yay connected")

client = mqtt.Client(client_id = "map_test")
client.on_connect = on_connect
# Blocking connect call
print("Attempting to connect...")
client.connect(host="tompop.tomseago.com", port=1883)


# Coordinates for the map_1080p.png ifle
lat_max_bound = lat_upper_left = Decimal('40.804337')
long_min_bound = long_upper_left = Decimal('-119.233532')
lat_min_bound = lat_bottom_right = Decimal('40.769762')
long_max_bound = long_bottom_right = Decimal('-119.152574')

lat_max_y = lat_max_bound - lat_min_bound
long_max_x = long_max_bound - long_min_bound

# From the 2023 KMZ dataset directly
ll_man = (Decimal('40.786393'), Decimal('-119.203515'))
temple = ll_temple = (Decimal('40.791255'), Decimal('-119.197142'))

# 4:00 and C 
ll_four_and_C = (Decimal("40.777589"), Decimal("-119.200443"))

# Previous test points from older image
test_1 = (Decimal('40.765757'), Decimal('-119.242022')) # test point at the upper left bound
test_2 = (Decimal('40.806824'), Decimal('-119.168263')) # test point at the bottom right bound

max_message_list_size = 20
new_gps_event_name = "<<NEW-GPS-COORDS>>"
new_gps_message_format = "Last BAAAHS Location @ {when}: [{lat}, {long}]"
print("lat bound", lat_max_y, "long bound", long_max_x)
class BaaahsMap:
    
    def __init__(self, image, icon, man):
        self.map_file = image
        self.icon_file = icon
        self.man_file = man
        self.root = tk.Tk()

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
        self.man_img = tk.PhotoImage(file=self.man_file, width=45, height=45)
        
        self.canvas = tk.Canvas(self.root, width=self.map_img.width(), height=self.map_img.height())
        self.canvas.create_image(0, 0, anchor='nw', image=self.map_img)

        # The Man
        man_pos = self.normalize_coords(ll_man[0], ll_man[1], self.man_img)
        self.canvas.create_image(man_pos[0], man_pos[1], anchor='c', image=self.man_img)

        # Initial BAAAHS
        self.baaahs_pos = self.normalize_coords(self.baaahs_pos_ll[0], self.baaahs_pos_ll[1], self.baaahs)
        self.baaahs_id = self.canvas.create_image(self.baaahs_pos[0], self.baaahs_pos[1], anchor='c', image=self.baaahs)

        self.canvas.event_add(new_gps_event_name, "None")
        self.canvas.bind(new_gps_event_name, self.new_baaahs_coords)

        # Right Text List
        text_box_width = self.map_img.width() / 6
        text_box_height = self.map_img.height() / 2
        self.text_box = tk.Listbox(self.root, bg='#000', font='arial', fg="#fff")
        self.text_box.pack()

        text_box_pos_x = text_box_width * 4
        text_box_pos_y = text_box_height
        self.canvas.create_window(text_box_pos_x, text_box_pos_y, anchor='se', window=self.text_box, width=text_box_width, height=text_box_height)
        #self.canvas.focus()

        self.text_box.insert(tk.END, "This is a message")
        self.text_box.insert(tk.END, "This is another message")
        self.text_box.insert(tk.END, "What would this do")
        self.text_box.focus()

        # Finally pack
        self.canvas.pack()

    def set_new_baaahs_cords(self, lat, long):
        self.baaahs_pos_ll = (Decimal(lat), Decimal(long))
        self.canvas.event_generate(new_gps_event_name)

    def new_baaahs_coords(self):
        self.baaahs_pos = self.normalize_coords(self.baaahs_pos_ll[0], self.baaahs_pos_ll[1], self.baaahs)
        self.canvas.delete(self.baaahs_id)
        self.baaahs_id = self.canvas.create_image(self.baaahs_pos[0], self.baaahs_pos[1], anchor='c', image=self.baaahs)
        self.add_message_to_box(new_gps_message_format.format(when=datetime.datetime.now(), lat=self.baaahs_pos_ll[0], long=self.baaahs_pos_ll[1]))
        # self.canvas.pack()
    
    def add_message_to_box(self, message):
        self.text_box.insert(tk.END, message)
        if self.text_box.size() > max_message_list_size:
            self.text_box.delete(0)
        
    
    def normalize_coords(self, lat, long, img):
        norm_lat = (lat - lat_min_bound)/lat_max_y
        norm_long = (long - long_min_bound)/long_max_x
        img_x = norm_long * self.bgMap_width
        img_y = norm_lat * self.bgMap_height

        img_offset_x = Decimal(img.width()) / Decimal(2)
        img_offset_y = Decimal(img.height()) / Decimal(2)

        # Verbose, but consistent
        output_x = img_x - img_offset_x
        output_y = img_y - img_offset_y

        # Invert the Y coordinate to make up for the differences in counting from 
        # "bottom to top" as latitude does versus "top to bottom" as most common
        # output displays do
        output_y = self.bgMap_height - output_y

        return (output_x, output_y)


MESSAGES_TOPIC = "bchat/rooms/main/*/messages"
GPS_TOPIC = "bchat/rooms/main/sheep_loc"

dir = os.getcwd()
#image = dir + "/bchat-map/resources/brc_map_better.pgm"
image = dir + "/bchat-map/resources/map_1080p.png"
icon = dir + "/bchat-map/resources/baaahs_icon.pgm"
man = dir + "/bchat-map/resources/the_man.pgm"

baaahsMap = BaaahsMap(image, icon, man)

def on_new_sheep_coords(client, userdata, message):
    new_lat = message.payload[0]
    new_long = message.payload[1]
    baaahsMap.set_new_baaahs_cords(new_lat, new_long)

message_format = "[{date}]:[{who}]: {text}"
def on_new_message(client, userdata, message):
    data = message.payload
    baaahsMap.add_message_to_box(message_format.format(date=data.Sent, who=data.From, text=data.Msg))

# Subs to GPS
client.subscribe(GPS_TOPIC, qos=0)
client.message_callback_add(GPS_TOPIC, on_new_sheep_coords)

# Subs to Messages
client.subscribe(MESSAGES_TOPIC, qos=0)
client.message_callback_add(MESSAGES_TOPIC, on_new_message)

baaahsMap.mainLoop()