#!/usr/bin/python3
from datetime import datetime
import requests
token = TOKEN
url = 'https://cloud-api.yandex.net/v1/disk/resources'


class Card:
    def __init__(self, name, url, size, date):
        self.name = name
        self.url = url
        self.size = size
        self.date = datetime.strptime(
                date,
                "%Y-%m-%dT%H:%M:%S%z").strftime("%Y.%m.%d(%H:%M:%S)")


def main():
    print("Choose file(s) to download 🠯: ")
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
    max_name = max(len(str(x['name'])) for x in items)
    max_size = max(len(str(x['size'])) for x in items)
    max_date = max(len(str(x['modified'])) for x in items)
    for item in items:
        card = Card(item['name'],
                    item['file'],
                    item['size'],
                    item['modified'])
        cards.append(card)
        print(f"[{len(cards)-1}]  {card.name.ljust(max_name)}  {str(card.size).ljust(max_size)}  {card.date.ljust(max_date)}")

    list_to_get = input(f"[enter number(s) (0..{len(cards)-1} | A -- for all)]: ")
    if list_to_get == 'A':
        list_to_get = list(range(len(cards)))
    else:
        list_to_get = [int(x) for x in list_to_get.split()]
        list_to_get = list(set(list_to_get))  # rm dublicates
    print(list_to_get)
    for i in list_to_get:
        try:
            response = requests.get(cards[i].url)
            open(cards[i].name, 'wb').write(response.content)
            print(cards[i].name, '✔')
        except requests.exceptions.RequestException as e:
            print("ERROR")
            print(e)
            print(cards[i].name, '✘')


if __name__ == "__main__":
    main()
