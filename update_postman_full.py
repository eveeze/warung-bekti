import json
import re

# Load the collection
with open('warung-backend.postman_collection.json', 'r') as f:
    collection = json.load(f)

# Helper to find a folder by name
def find_folder(items, name):
    for item in items:
        if item.get('name') == name:
            return item
    return None

# Helper to create a request item
def create_request(name, method, url_raw, body_raw=None, test_script=None):
    item = {
        "name": name,
        "request": {
            "method": method,
            "header": [],
            "url": {
                "raw": url_raw,
                "host": ["{{base_url}}"],
                "path": url_raw.replace("{{base_url}}/", "").split("/")
            }
        }
    }
    
    # Extract query params if any
    if "?" in url_raw:
        base, query = url_raw.split("?")
        item["request"]["url"]["raw"] = url_raw
        item["request"]["url"]["path"] = base.replace("{{base_url}}/", "").split("/")
        queries = []
        for q in query.split("&"):
            k, v = q.split("=")
            queries.append({"key": k, "value": v})
        item["request"]["url"]["query"] = queries

    if body_raw:
        item["request"]["body"] = {
            "mode": "raw",
            "raw": body_raw,
            "options": {"raw": {"language": "json"}}
        }
    
    if test_script:
        item["event"] = [{
            "listen": "test",
            "script": {
                "exec": test_script,
                "type": "text/javascript"
            }
        }]
    
    return item

# --- 1. Fix Admin Flow -> Customers: Add "Create Debt Transaction" ---
admin_flow = find_folder(collection['item'], "1. Admin Flow (All Requests)")
if admin_flow:
    customers_folder = find_folder(admin_flow['item'], "Customers")
    if customers_folder:
        # Check if already exists to avoid dupes
        exists = any(i['name'] == "Create Debt Transaction" for i in customers_folder['item'])
        if not exists:
            # Find index of Record Kasbon Payment
            idx = -1
            for i, item in enumerate(customers_folder['item']):
                if item['name'] == "Record Kasbon Payment":
                    idx = i
                    break
            
            if idx != -1:
                # Create the request
                req = create_request(
                    "Create Debt Transaction",
                    "POST",
                    "{{base_url}}/api/v1/transactions",
                    json.dumps({
                        "items": [{"product_id": "{{product_id_1}}", "quantity": 2}],
                        "payment_method": "kasbon",
                        "customer_id": "{{customer_id}}",
                        "paid_amount": 0,
                        "notes": "Test Kasbon Payment Admin"
                    }, indent=4),
                    [
                        "pm.test(\"Status code is 201\", function () {",
                        "    pm.response.to.have.status(201);",
                        "});",
                        "var jsonData = pm.response.json();",
                        "if(jsonData.data && jsonData.data.id) {",
                        "    pm.environment.set(\"transaction_kasbon_id\", jsonData.data.id);",
                        "}"
                    ]
                )
                customers_folder['item'].insert(idx, req)
                print("Added 'Create Debt Transaction' to Admin Flow")

# --- 2. Fix Refillable Flow -> Adjust Body ---
if admin_flow: # Assuming Refillables might be in Admin Flow or logic is shared
    pass # Actually Refillable Adjust in the user snippet was in "3. Inventory Flow" usually?
    # Let's search globally for "Adjust Container Stock"
    
def fix_refillable_adjust(items):
    for item in items:
        if item.get('name') == "Adjust Container Stock":
            # Fix body
            if 'request' in item and 'body' in item['request']:
                try:
                    body = json.loads(item['request']['body']['raw'])
                    body['container_id'] = "{{container_id}}" # Ensure variable
                    body['empty_change'] = 1
                    body['full_change'] = -1
                    item['request']['body']['raw'] = json.dumps(body, indent=4)
                    print(f"Fixed Refillable Adjust Body in {item['name']}")
                except:
                    pass
        if 'item' in item:
            fix_refillable_adjust(item['item'])

fix_refillable_adjust(collection['item'])

# --- 3. Add Missing Transaction Routes to "1. Admin Flow" -> "Transactions" ---
if admin_flow:
    tx_folder = find_folder(admin_flow['item'], "Transactions")
    if not tx_folder:
        # Create if missing
        tx_folder = {"name": "Transactions", "item": []}
        admin_flow['item'].append(tx_folder)
    
    # 3a. GET /transactions
    if not any(i['name'] == "List Transactions" for i in tx_folder['item']):
        tx_folder['item'].append(create_request(
            "List Transactions",
            "GET",
            "{{base_url}}/api/v1/transactions?page=1&per_page=10&start_date=2024-01-01&end_date=2025-12-31",
            None,
            ["pm.test(\"Status code is 200\", function () { pm.response.to.have.status(200); });"]
        ))
        print("Added 'List Transactions'")

    # 3b. POST /transactions/calculate
    if not any(i['name'] == "Calculate Transaction" for i in tx_folder['item']):
        tx_folder['item'].insert(0, create_request( # Put at top
            "Calculate Transaction",
            "POST",
            "{{base_url}}/api/v1/transactions/calculate",
            json.dumps({
                "items": [{"product_id": "{{product_id_1}}", "quantity": 5}],
                "payment_method": "cash"
            }, indent=4),
            ["pm.test(\"Status code is 200\", function () { pm.response.to.have.status(200); });"]
        ))
        print("Added 'Calculate Transaction'")
        
    # 3c. GET /transactions/{id}
    if not any(i['name'] == "Get Transaction Detail" for i in tx_folder['item']):
        tx_folder['item'].append(create_request(
            "Get Transaction Detail",
            "GET",
            "{{base_url}}/api/v1/transactions/{{transaction_kasbon_id}}", # Use one we know exists from Debt test
            None,
            ["pm.test(\"Status code is 200\", function () { pm.response.to.have.status(200); });"]
        ))
        print("Added 'Get Transaction Detail'")

# --- 4. Validate Environment Variables in Scripts ---
# Just a heuristic check or simple fix for now, ensuring create transaction captures ID
test_script_create_tx = [
    "pm.test(\"Status code is 201\", function () {",
    "    pm.response.to.have.status(201);",
    "});",
    "",
    "var jsonData = pm.response.json();",
    "if (jsonData.data && jsonData.data.id) {",
    "    pm.environment.set(\"transaction_id\", jsonData.data.id);",
    "    // Also set for Cancel/Refund tests if needed",
    "    pm.environment.set(\"transaction_kasbon_id\", jsonData.data.id);",
    "}",
    "if (jsonData.data && jsonData.data.items && jsonData.data.items.length > 0) {",
    "    pm.environment.set(\"tx_item_id\", jsonData.data.items[0].id);",
    "}"
]

def update_create_tx_scripts(items):
    for item in items:
        if item.get('name') == "Create Transaction":
            item['event'] = [{
                "listen": "test",
                "script": {
                    "exec": test_script_create_tx,
                    "type": "text/javascript"
                }
            }]
            print(f"Updated script for {item['name']}")
        if 'item' in item:
            update_create_tx_scripts(item['item'])

update_create_tx_scripts(collection['item'])

# Save
with open('warung-backend.postman_collection.json', 'w') as f:
    json.dump(collection, f, indent=4)

print("Postman collection updated successfully.")
