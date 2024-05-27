import os
import sys
import re


def update_elasticsearch_enabled(lines, section, identifier) -> []:
    updated_lines = []
    section_found = False

    for line in lines:
        if line.startswith("[" + section + "]"):
            section_found = True
        if section_found and identifier in line:
            line = "    Enabled           = true\n"
            section_found = False
        updated_lines.append(line)

    return updated_lines


def main():
    sovereign_path = sys.argv[1]

    toml_path = sovereign_path + "/proxy/config/external.toml"

    with open(toml_path, 'r') as file:
        lines = file.readlines()

    updated_lines = update_elasticsearch_enabled(lines, "ElasticSearchConnector", "Enabled")

    with open(toml_path, 'w') as file:
        file.writelines(updated_lines)


if __name__ == "__main__":
    main()
