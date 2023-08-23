import tkinter as tk
from tkinter import PhotoImage
from decimal import Decimal
import os

# Old coordinates for brc_map_better.pgm
# lat_max_bound = Decimal('40.806824')
# lat_min_bound = Decimal('40.765757')
# long_max_bound = Decimal('-119.168263')
# long_min_bound = Decimal('-119.242022')

# Coordinates for the map_1080p.png ifle
lat_max_bound = lat_upper_left = Decimal('40.804337')
long_min_bound = long_upper_left = Decimal('-119.233532')
lat_min_bound = lat_bottom_right = Decimal('40.769762')
long_max_bound = long_bottom_right = Decimal('-119.152574')

lat_max_y = lat_max_bound - lat_min_bound
long_max_x = long_max_bound - long_min_bound
# man_lat = Decimal('40.786400')
# man_long = Decimal('-119.203500')

# Previous coordinate for temple
#temple = (Decimal('40.791346'), Decimal('-119.200243'))

# From the 2023 KMZ dataset directly
ll_man = (Decimal('40.786393'), Decimal('-119.203515'))
temple = ll_temple = (Decimal('40.791255'), Decimal('-119.197142'))

# Previous test points from older image
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
        print("bgMap_width=",self.bgMap_width,"  bgMap_height=",self.bgMap_height)

        self.baaahs = tk.PhotoImage(file=self.icon_file, width=40, height=34)
        self.man_img = tk.PhotoImage(file=self.man_file, width=45, height=45)
        
        self.canvas = tk.Canvas(self.root, width=self.map_img.width(), height=self.map_img.height())
        self.canvas.create_image(0, 0, anchor='nw', image=self.map_img)
        ## self.canvas.create_image(50,50, anchor='nw', image=self.baaahs)
        man_pos = self.normalize_coords(ll_man[0], ll_man[1], self.man_img)
        man_x_pos = man_pos[0]
        man_y_pos = man_pos[1]
        # self.new_gps_coord(40.78652, -119.206567, self.baaahs)
        #self.canvas.create_image(man_y_pos, man_x_pos, anchor='nw', image=self.man_img)
        self.canvas.create_image(man_x_pos, man_y_pos, anchor='c', image=self.man_img)
        
        # test_1_pos = self.normalize_coords(test_1[0], test_1[1], self.baaahs)
        # self.canvas.create_image(test_1_pos[0], test_1_pos[1], anchor='nw', image=self.baaahs)

        # test_2_pos = self.normalize_coords(test_2[0], test_2[1], self.baaahs)
        # self.canvas.create_image(test_2_pos[0], test_2_pos[1], anchor='nw', image=self.baaahs)

        temple_pos = self.normalize_coords(temple[0], temple[1], self.baaahs)
        self.canvas.create_image(temple_pos[0], temple_pos[1], anchor='c', image=self.baaahs)

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
        print("img_y=", img_y, "  img_x=",img_x)
        img_offset_x = Decimal(img.width()) / Decimal(2)
        img_offset_y = Decimal(img.height()) / Decimal(2)
        print("offset x", img_offset_x, "offset y", img_offset_y)

        # Verbose, but consistent
        output_x = img_x - img_offset_x
        output_y = img_y - img_offset_y

        # Invert the Y coordinate to make up for the differences in counting from 
        # "bottom to top" as latitude does versus "top to bottom" as most common
        # output displays do
        output_y = self.bgMap_height - output_y

        return (output_x, output_y)



dir = os.getcwd()
#image = dir + "/bchat-map/resources/brc_map_better.pgm"
image = dir + "/bchat-map/resources/map_1080p.png"
icon = dir + "/bchat-map/resources/baaahs_icon.pgm"
man = dir + "/bchat-map/resources/the_man.pgm"
BaaahsMap(image, icon, man)
