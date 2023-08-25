import serial               
from time import sleep
import sys                  
import paho.mqtt.client as mqtt
from decimal import Decimal

def GPS_Info():
    global NMEA_buff
    global lat_in_degrees
    global long_in_degrees
    nmea_time = []
    nmea_latitude = []
    nmea_longitude = []
    nmea_time = NMEA_buff[0]                    #extract time from GPGGA string
    nmea_latitude = NMEA_buff[1]                #extract latitude from GPGGA string
    north_south = NMEA_buff[2]
    east_west = NMEA_buff[4]                                  
    nmea_longitude = NMEA_buff[3]               #extract longitude from GPGGA string
    
    print("NMEA Time: ", nmea_time,'\n')
    print ("NMEA Latitude:", nmea_latitude,"NMEA Longitude:", nmea_longitude,'\n')
    
    try:
        lat = float(nmea_latitude)                  #convert string into float for calculation
        longi = float(nmea_longitude)               #convertr string into float for calculation
        
        lat_in_degrees = convert_to_degrees(lat, north_south)    #get latitude in degree decimal format
        long_in_degrees = convert_to_degrees(longi, east_west) #get longitude in degree decimal format
    except:
        print("no data yet")
        sleep(500)
    
#convert raw NMEA string into degree decimal format   
def convert_to_degrees(raw_value, hemis):
    neg = hemis == 'S' or hemis == 'W' 
    decimal_value = raw_value/100.00
    degrees = int(decimal_value)
    mm_mmmm = (decimal_value - int(decimal_value))/0.6
    position = degrees + mm_mmmm
    position = "%.4f" %(position)
    
    if (neg):
        position = '-' + position
        
    return position
    


# NEO M9 stuff
baud_rate = 38400

gpgga_info = "$GNGGA,"
ser = serial.Serial ("/dev/ttyAMA0", baud_rate)
GPGGA_buffer = 0
NMEA_buff = 0
lat_in_degrees = 0
long_in_degrees = 0

# Parse command line args for fun and profit (and to make development life easier)
import argparse

parser = argparse.ArgumentParser(description="BAAAHS Chat map display")
parser.add_argument("--mqtt_host", default="localhost", type=str, help="MQTT host address")
parser.add_argument("--mqtt_port", default=1883, type=int, help="MQTT port number")
parser.add_argument("--mqtt_id", default="gps_test", type=str, help="MQTT client ID")
args = parser.parse_args()

# MQTT stuff
def on_connect(client, userdata, flags, rc):
    print("Yay gps connected")

client = mqtt.Client(client_id = args.mqtt_id)
client.connect(host=args.mqtt_host, port=args.mqtt_port)
client.on_connect = on_connect
client.loop_start()

min_diff = Decimal("0.00005")
lat_long_message = "{lat:.7f},{long:.7f}"
ll_four_and_C = (Decimal("40.777589"), Decimal("-119.200443"))
last_ll = ll_four_and_C

try:
    while True:
        received_data = (str)(ser.readline())                   #read NMEA string received
        GPGGA_data_available = received_data.find(gpgga_info)   #check for NMEA GPGGA string                 
        if (GPGGA_data_available>0):
            GPGGA_buffer = received_data.split(gpgga_info,1)[1]  #store data coming after "$GPGGA," string 
            NMEA_buff = (GPGGA_buffer.split(','))               #store comma separated data in buffer
            GPS_Info()                                          #get time, latitude, longitude

            if (lat_in_degrees == 0 or long_in_degrees == 0):
                continue
            
            new_ll = (Decimal(lat_in_degrees), Decimal(long_in_degrees))
            if ((min_diff < abs(last_ll[0] - new_ll[0])) or min_diff < abs(last_ll[1] - new_ll[1])):
                last_ll = new_ll
                msg_str = lat_long_message.format(lat=last_ll[0], long=last_ll[1])
                client.publish("bchat/rooms/main/sheep_loc", msg_str)   

            print("lat in degrees:", lat_in_degrees," long in degree: ", long_in_degrees, '\n')
        
       
except KeyboardInterrupt:
    sys.exit(0)