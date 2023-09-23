#!/usr/bin/python3
from datetime import datetime
import requests
import os
import sys
import itertools

token = os.environ.get("TOKEN")
url = 'https://cloud-api.yandex.net/v1/disk/resources'
symbs = ['🌍', '🌏', '🌎']
# symbs = ['🕛',
#  '🕒',
#  '🕕',
#  '🕘']
chars = itertools.cycle(symbs)


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
        print(f"[{str(len(cards)-1).rjust(n)}]",
              f"{card.name.ljust(max_name)}",
              f"{str(card.size).ljust(max_size)}",
              f"{card.date.ljust(max_date)}")

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
            with open(cards[i].name, 'wb') as file:
                for data in response.iter_content(1024):
                    sys.stdout.write('\r{}{}'.format(cards[i].name,
                                                     next(chars)))
                    sys.stdout.flush()
                    file.write(data)
                sys.stdout.write('\r{} ✔\n'.format(cards[i].name))
                sys.stdout.flush()
        except requests.exceptions.RequestException as e:
            print("ERROR")
            print(e)
            print(cards[i].name, '✘\n')


if __name__ == "__main__":
    main()
