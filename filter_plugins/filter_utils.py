import os

def join_paths(a, *p):
    return os.path.join(a, *p)

class FilterModule(object):
    def filters(self):
        return {
            'join_paths': join_paths
        }
