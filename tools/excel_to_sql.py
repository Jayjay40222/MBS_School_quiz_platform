import pandas as pd
import sqlite3
import os

conn = sqlite3.connect("users.db")

files = ["Grade_9D.xlsx", "Grade_9C.xlsx", "Grade_9B.xlsx", "Grade_9A.csv"]

for file in files:
    # Read file based on type
    if file.endswith(".csv"):
        df = pd.read_csv(file)
    else:
        df = pd.read_excel(file)

    # Rename columns to match your DB schema
    df = df.rename(columns={
        "Student_Name": "full_name",
        "Class": "class",
        "Password": "password"
    })

    # Keep only needed columns
    df = df[["ID", "full_name",  "password", "class"]]

    # Insert into existing table
    df.to_sql("users", conn, if_exists="append", index=False)

conn.close()