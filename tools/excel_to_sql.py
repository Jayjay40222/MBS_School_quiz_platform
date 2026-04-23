import pandas as pd
import sqlite3

conn = sqlite3.connect("users.db")
cursor = conn.cursor()

files = ["Grade_12A.xlsx", "Grade_12B.xlsx"]

def clean(val):
    if pd.isna(val):
        return None
    return str(val).strip()

total_inserted = 0
total_updated = 0
failed = 0

for file in files:

    # ✔ correct file handling (NO nested loop)
    if file.endswith(".csv"):
        df = pd.read_csv(file)
    else:
        df = pd.read_excel(file)

    df.columns = df.columns.str.strip()

    df = df.rename(columns={
        "Student_Name": "full_name",
        "Class": "class",
        "Password": "password",
        "ID": "studID"
    })

    df = df[["studID", "full_name", "password", "class"]]

    df["studID"] = df["studID"].apply(clean)
    df["full_name"] = df["full_name"].apply(clean)
    df["password"] = df["password"].apply(clean)
    df["class"] = df["class"].apply(clean)

    df = df.dropna(subset=["studID", "full_name", "password"])

    for _, row in df.iterrows():
        try:
            cursor.execute("""
                INSERT INTO users (studID, full_name, password, class, points)
                VALUES (?, ?, ?, ?, 0)
                ON CONFLICT(studID) DO UPDATE SET
                    full_name = excluded.full_name,
                    password = excluded.password,
                    class = excluded.class
            """, (
                row["studID"],
                row["full_name"],
                row["password"],
                row["class"]
            ))

            if cursor.rowcount == 1:
                total_inserted += 1
            else:
                total_updated += 1

        except Exception as e:
            failed += 1
            print("FAILED ROW:", row.to_dict(), "ERROR:", e)

conn.commit()
conn.close()

print("Done importing")
print("Inserted:", total_inserted)
print("Updated:", total_updated)
print("Failed:", failed)