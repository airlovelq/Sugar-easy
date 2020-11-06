from PIL import Image
import csv
import tempfile
import zipfile
import os
import io
import pandas as pd
import numpy as np

def load_image_classification_dataset(dataset_path, mode='RGB'):
    # Create temp directory to unzip to
    # with tempfile.TemporaryDirectory() as d:
    #     dataset_zipfile = zipfile.ZipFile(dataset_path, 'r')
    #     dataset_zipfile.extractall(path=d)
    #     dataset_zipfile.close()

    # Read images.csv, and read image paths & classes
    image_paths = []
    image_classes = []
    images_csv_path = os.path.join(dataset_path, 'images.csv')
    try:
        with open(images_csv_path, mode='r') as f:
            reader = csv.DictReader(f)
            (image_paths, image_classes) = zip(*[(row['path'], int(row['class'])) for row in reader])
    except Exception as e:
        # traceback.print_stack()
        raise Exception()

    # Load images from files
    full_image_paths = [os.path.join(dataset_path, x) for x in image_paths]
    pil_images = _load_pil_images(full_image_paths, mode=mode)

    num_classes = len(set(image_classes))
    num_samples = len(image_paths)

    return (pil_images, image_classes, num_samples, num_classes)

def _load_pil_images(image_paths, mode='RGB'):
    pil_images = []
    for image_path in image_paths:
        with open(image_path, 'rb') as f:
            encoded = io.BytesIO(f.read())
            pil_image = Image.open(encoded).convert(mode)
            pil_images.append(pil_image)
    
    return pil_images

def load_table_regression_dataset(dataset_path):
    pass

def load_table_classification_dataset(dataset_path, feature_list, target_list):
    preader = pd.read_csv(dataset_path)
    # schema_list = preader.columns.tolist()
        #all row
    features = preader[feature_list]
    classes = preader[target_list]
    # classes = np.asarray(preader[target_list].to_numpy().reshape(-1), dtype='int')
    num_classes = len(set(np.asarray(preader[target_list].to_numpy().reshape(-1), dtype='int')))
    return features, classes, len(features), num_classes
