#!/usr/bin/python3
from pprint import pprint
from datetime import datetime
import sys

import urllib.parse as ul
import requests
token = TOKEN
url = 'https://cloud-api.yandex.net/v1/disk/resources'

class Card:
    def __init__(self, name, url, size, date):
        self.name = name
        self.url = url
        self.size = size
        self.date = datetime.strptime(date, "%Y-%m-%dT%H:%M:%S%z").strftime("%Y.%m.%d(%H:%M:%S)")
    
def main():
    print("Choose file(s) to download: ")
    headers = {'Authorization': 'OAuth ' + token}
    params = {'path': '/kindle'}
    try:
        response = requests.get(url, params=params, headers=headers)
    except requests.exceptions.RequestException as e:
        print("ERROR")
        print(e)

    cards = []
    items = response.json()['_embedded']['items']
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

    list_to_get = input(f"[enter number(s) (0..{len(cards)-1})]: ")
    print(list_to_get)

if __name__ == "__main__":
    main()
# curl -H "Authorization: OAuth $token" https://cloud-api.yandex.net/v1/disk/resources?path=%2Fkindle | jq '._embedded.items[].name'
# wget $(curl -s -H "Authorization: OAuth $token" https://cloud-api.yandex.net/v1/disk/resources/download?path=%2Fkindle%2F${file} | jq -r '.href') -O ${file}
