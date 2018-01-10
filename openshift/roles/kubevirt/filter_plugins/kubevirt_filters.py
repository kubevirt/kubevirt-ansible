import yaml
from collections import defaultdict

from ansible.errors import AnsibleError

try:
    from __main__ import display
except ImportError:
    from ansible.utils.display import Display
    display = Display()

def mf_to_dict(mf):
    """
    Split Kubevirt manifest into a dict:
    kind -> list of dicts

    Args:
        mf (str): Path to a yaml file

    Returns:
        (dict)
    """
    d = defaultdict(list)

    with open(mf) as f:
        docs = yaml.safe_load_all(f)

        for doc in docs:
            kind = doc['kind']
            name = doc['metadata']['name']
            d[kind].append(
                {
                    'name': name,
                    'manifest': doc
                }
            )

    return dict(d)

class FilterModule(object):
    def filters(self):
        return {
            'mf_to_dict': mf_to_dict
        }
