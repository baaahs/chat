import tkinter as tk
from tkinter import PhotoImage
import os

lat_max_bound = 40.806251
lat_min_bound = 40.768335
long_max_bound = -119.160021
long_min_bound = -119.241827
lat_max_y = lat_max_bound - lat_min_bound
long_max_x = long_max_bound - long_min_bound
man_lat = 40.786443
man_long = -119.206610

class BaaahsMap:
    
    def __init__(self, image, icon, man):
        self.map_file = image
        self.icon_file = icon
        self.man_file = man
        self.root = tk.Tk()
        self.widgets()
        self.x_baaahs_pos = 0
        self.y_baaahs_pos = 0 

        self.root.mainloop()

    def widgets(self):
        self.map_img = tk.PhotoImage(file=self.map_file)
        self.bgMap_width = self.map_img.width()
        self.bgMap_height = self.map_img.height()
        self.baaahs = tk.PhotoImage(file=self.icon_file, width=40, height=34)
        self.man_img = tk.PhotoImage(file=self.man_file, width=45, height=45)
        
        self.canvas = tk.Canvas(self.root, width=self.map_img.width(), height=self.map_img.height())
        self.canvas.create_image(0, 0, anchor='nw', image=self.map_img)
        self.canvas.create_image(50,50, anchor='nw', image=self.baaahs)
        man_pos = self.normalize_coords(man_lat, man_long, self.man_img)
        man_x_pos = man_pos[0]
        man_y_pos = man_pos[1]
        self.new_gps_coord(40.78652, -119.206567, self.baaahs)
        self.canvas.create_image(man_y_pos, man_x_pos, anchor='nw', image=self.man_img)

        self.canvas.pack()
    
    def new_gps_coord(self, lat, long, img):
        new_pos = self.normalize_coords(lat, long, img)
        self.canvas.create_image(new_pos[0], new_pos[1], anchor='nw', image=self.baaahs)
        self.canvas.pack()
    
    def normalize_coords(self, lat, long, img):
        norm_lat = ((lat - lat_min_bound)/lat_max_y) * self.bgMap_width
        norm_long = ((long - long_min_bound)/long_max_x) * self.bgMap_height
        img_offset_x = img.width() / 2
        img_offset_y = img.height() / 2
        return (norm_long - img_offset_x, norm_lat - img_offset_y)



dir = os.getcwd()
image = dir + "/bchat-map/resources/brc_map.png"
icon = dir + "/bchat-map/resources/baaahs_icon.pgm"
man = dir + "/bchat-map/resources/the_man.pgm"
BaaahsMap(image, icon, man)
