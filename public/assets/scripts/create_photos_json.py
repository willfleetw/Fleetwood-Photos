from importlib.resources import path
import os
import glob
import pathlib
from PIL import Image, ExifTags
import json

dir_name = 'C:\\src\\Fleetwood-Photos\\public\\images\\small\\'

list_of_files = filter(os.path.isfile, glob.glob(dir_name + '*'))

list_of_files = sorted(list_of_files, key=os.path.getmtime, reverse=True)
image_files = []

for file in list_of_files:
  if not file.endswith('.jpg'):
    continue
  image_files.append({'name': os.path.basename(file)})

with open('C:\\src\Fleetwood-Photos\\public\\data\\photos.json', 'w') as write_file:
  json.dump(image_files, write_file, indent=2)