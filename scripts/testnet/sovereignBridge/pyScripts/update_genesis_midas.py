import os
import sys
import re
import json


def read_and_concatenate_addresses(file_path):
    with open(file_path, 'r') as file:
        data = json.load(file)
        addresses = [f"{node['address']}" for node in data]
        return addresses


def main():
    sovereign_path = sys.argv[1]

    file_path = os.path.join(sovereign_path, 'node/config/genesis.json')
    addresses = read_and_concatenate_addresses(os.path.expanduser(file_path))
    print(addresses)

    current_path = os.getcwd()
    project = 'mx-chain-go'
    index = current_path.find(project)
    project_path = current_path[:index + len(project)]
    genesis_path = project_path + "/cmd/sovereignnode/config/genesis.json"

    with open(genesis_path, 'r') as file:
        genesis_data = json.load(file)

    for index, address in enumerate(addresses):
        genesis_data[index]['address'] = address


    with open(file_path, 'w') as file:
        json.dump(genesis_data, file, ensure_ascii=False, indent=2)


if __name__ == "__main__":
    main()
