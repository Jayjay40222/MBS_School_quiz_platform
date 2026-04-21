import pandas as pd
import sqlite3

conn = sqlite3.connect("users.db")
cursor = conn.cursor()

files = ["Grade_12A.xlsx", "Grade_12B.xlsx"]

for file in files:
    if file.endswith(".csv"):
        df = pd.read_csv(file)
    else:
        df = pd.read_excel(file)

    # Normalize column names
    df = df.rename(columns={
        "Student_Name": "full_name",
        "Class": "class",
        "Password": "password",
        "ID": "studID"
    })

    # Keep required columns
    df = df[["studID", "full_name", "password", "class"]]

    # Remove incomplete rows
    df = df.dropna(subset=["studID", "full_name", "password"])

    # Add points column (default 0)
    df["points"] = 0

    # Insert safely
    for _, row in df.iterrows():
        cursor.execute("""
            INSERT OR IGNORE INTO users (studID, full_name, password, class, points)
            VALUES (?, ?, ?, ?, ?)
        """, (
            row["studID"],
            row["full_name"],
            row["password"],
            row["class"],
            row["points"]
        ))

conn.commit()
conn.close()