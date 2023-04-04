import requests
import time

order_response = requests.post('http://localhost:8787/order')
order_data = order_response.json()
order_id = order_data['orderid']

print("Order ID:", order_id)

headers = {
    'Content-Type': 'application/x-www-form-urlencoded',
}

while True:
    data = '{"orderid": "' + order_id + '"}'
    status_response = requests.post('http://localhost:8787/status', headers=headers, data=data)
    status_data = status_response.json()
    status = status_data['status']
    print("Status:", status)
    if status == 'delivered':
        break
    time.sleep(10)
