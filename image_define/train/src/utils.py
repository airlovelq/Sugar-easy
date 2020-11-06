from importlib import import_module

def load_model_class(module, model_class):
    clazz = None
    try:
        # Import model file as module
        mod = import_module(module)
        # Extract model class from module
        clazz = getattr(mod, model_class)
    except Exception as e:
        raise e
        # logger.log(logging.INFO, 'Getting Model Class Failed')
        # logger.log(logging.ERROR, repr(e))
        # raise e
    finally:
        pass
    return clazz