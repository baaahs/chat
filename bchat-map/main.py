import tkinter as tk
from tkinter import PhotoImage
from decimal import Decimal
import os

lat_max_bound = Decimal('40.806824')
lat_min_bound = Decimal('40.765757')
long_max_bound = Decimal('-119.168263')
long_min_bound = Decimal('-119.242022')
lat_max_y = lat_max_bound - lat_min_bound
long_max_x = long_max_bound - long_min_bound
man_lat = Decimal('40.786400')
man_long = Decimal('-119.203500')

temple = (Decimal('40.791346'), Decimal('-119.200243'))

test_1 = (Decimal('40.765757'), Decimal('-119.242022')) # test point at the upper left bound
test_2 = (Decimal('40.806824'), Decimal('-119.168263')) # test point at the bottom right bound

print("lat bound", lat_max_y, "long bound", long_max_x)
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
        self.bgMap_width = Decimal(self.map_img.width())
        self.bgMap_height = Decimal(self.map_img.height())
        self.baaahs = tk.PhotoImage(file=self.icon_file, width=40, height=34)
        self.man_img = tk.PhotoImage(file=self.man_file, width=45, height=45)
        
        self.canvas = tk.Canvas(self.root, width=self.map_img.width(), height=self.map_img.height())
        self.canvas.create_image(0, 0, anchor='nw', image=self.map_img)
        ## self.canvas.create_image(50,50, anchor='nw', image=self.baaahs)
        man_pos = self.normalize_coords(man_lat, man_long, self.man_img)
        man_x_pos = man_pos[0]
        man_y_pos = man_pos[1]
        # self.new_gps_coord(40.78652, -119.206567, self.baaahs)
        self.canvas.create_image(man_y_pos, man_x_pos, anchor='nw', image=self.man_img)
        
        test_1_pos = self.normalize_coords(test_1[0], test_1[1], self.baaahs)
        self.canvas.create_image(test_1_pos[0], test_1_pos[1], anchor='nw', image=self.baaahs)

        test_2_pos = self.normalize_coords(test_2[0], test_2[1], self.baaahs)
        self.canvas.create_image(test_2_pos[0], test_2_pos[1], anchor='nw', image=self.baaahs)

        temple_pos = self.normalize_coords(temple[0], temple[1], self.baaahs)
        self.canvas.create_image(temple_pos[0], temple_pos[1], anchor='nw', image=self.baaahs)

        self.canvas.pack()
    
    def new_gps_coord(self, lat, long, img):
        new_pos = self.normalize_coords(lat, long, img)
        self.canvas.create_image(new_pos[0], new_pos[1], anchor='nw', image=self.baaahs)
        self.canvas.pack()
    
    def normalize_coords(self, lat, long, img):
        norm_lat = (lat - lat_min_bound)/lat_max_y
        norm_long = (long - long_min_bound)/long_max_x
        img_x = norm_long * self.bgMap_width
        img_y = norm_lat * self.bgMap_height
        print("norm lat (y)", norm_lat, "norm long (x)", norm_long)
        img_offset_x = Decimal(img.width()) / Decimal(2)
        img_offset_y = Decimal(img.height()) / Decimal(2)
        print("offset x", img_offset_x, "offset y", img_offset_y)
        return (img_x - img_offset_x, img_y - img_offset_y)



dir = os.getcwd()
image = dir + "/bchat-map/resources/brc_map_better.pgm"
icon = dir + "/bchat-map/resources/baaahs_icon.pgm"
man = dir + "/bchat-map/resources/the_man.pgm"
BaaahsMap(image, icon, man)
