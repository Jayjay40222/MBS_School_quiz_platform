import pandas as pd
import sqlite3

files = ["Grade_9A.csv", "Grade_9B.xlsx", "Grade_9C.xlsx", "Grade_9D.xlsx"]

excel_ids = set()

for file in files:
    if file.endswith(".csv"):
        df = pd.read_csv(file)
    else:
        df = pd.read_excel(file)

    df = df.rename(columns={"ID": "studID"})
    df["studID"] = df["studID"].astype(str).str.strip()
    excel_ids.update(df["studID"].dropna())

conn = sqlite3.connect("users.db")
cursor = conn.cursor()

cursor.execute("SELECT studID FROM users")
db_ids = set(row[0] for row in cursor.fetchall())

extra_in_db = db_ids - excel_ids

print("Extra in DB:", extra_in_db)