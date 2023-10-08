#!/usr/bin/python3
from datetime import datetime
import requests
import os
import sys
import itertools


class TerminalColors:
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RESET = '\033[0m'


token = os.environ.get("TOKEN")
url = 'https://cloud-api.yandex.net/v1/disk/resources'
symbs = ['◐', '◓', '◑', '◒']
chars = itertools.cycle(symbs)
term_cols = 37


def cut_str(string):
    if len(string) > term_cols:
        strr = string[:term_cols] + '...'
        return strr
    else:
        return string 


def human_size(nbytes):
    suffixes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    i = 0
    while nbytes >= 1024 and i < len(suffixes)-1:
        nbytes /= 1024.
        i += 1
    f = ('%.2f' % nbytes).rstrip('0').rstrip('.')
    return '%s %s' % (f, suffixes[i])


class Card:
    def __init__(self, name, url, size, date):
        self.name = name
        self.url = url
        self.size = human_size(size)
        self.date = datetime.strptime(
                date,
                "%Y-%m-%dT%H:%M:%S%z").strftime("%d.%m.%y(%H:%M)")


def main():
    print("Choose file(s) to download 🃏: ")
    headers = {'Authorization': 'OAuth ' + token}
    params = {'path': '/kindle'}
    try:
        response = requests.get(url, params=params, headers=headers)
    except requests.exceptions.RequestException as e:
        print("ERROR")
        print(e)

    cards = []
    items = response.json()['_embedded']['items']

    # for pretty output
    n = len(str(len(items)))
    max_name = max(len(str(x['name'])) for x in items)
    max_size = max(len(str(human_size(x['size']))) for x in items)
    max_date = max(len(str(x['modified'])) for x in items)
    for item in items:
        card = Card(item['name'],
                    item['file'],
                    item['size'],
                    item['modified'])
        cards.append(card)
        long_name = cut_str(card.name)
        print(f"[{str(len(cards)-1).rjust(n)}]",
              # f"{card.name.ljust(max_name)}",
              f"{long_name.ljust(term_cols+3)}",
              f"{str(card.size).ljust(max_size)}",
              f"{card.date.ljust(max_date)}".rstrip())

    list_to_get = input(f"\n[enter number(s) (0..{len(cards)-1}|A)]: ")
    if list_to_get == 'A':
        list_to_get = list(range(len(cards)))
    else:
        list_to_get = [int(x) for x in list_to_get.split()]
        list_to_get = list(set(list_to_get))  # rm dublicates
    print(list_to_get)
    for i in list_to_get:
        try:
            response = requests.get(cards[i].url, stream=True)
            size = human_size(int(response.headers.get('Content-Length', 0)))
            cur_size = 0
            with open(cards[i].name, 'wb') as file:
                for data in response.iter_content(1024):
                    cur_size += len(data)
                    str_to_flash = '\r{}{} {} [{}/{}]'.format(
                        TerminalColors.YELLOW,
                        next(chars),
                        cut_str(cards[i].name),
                        human_size(cur_size),
                        size)
                    sys.stdout.write(str_to_flash)
                    sys.stdout.flush()
                    file.write(data)
                sys.stdout.write('\r' + ' ' * len(str_to_flash))
                sys.stdout.write('\r{}✔ {}\n'.format(TerminalColors.GREEN,
                                                     cut_str(cards[i].name)))
                sys.stdout.flush()
                sys.stdout.write(TerminalColors.RESET)
                sys.stdout.flush()
        except requests.exceptions.RequestException as e:
            print("ERROR")
            print(e)
            print(cards[i].name, '✘\n')


if __name__ == "__main__":
    main()
