from app import create_app
from flask import render_template, session
import os
import logging

app = create_app()

# Ensure sessions work when templates read session
app.secret_key = os.environ.get("SECRET_KEY", "dev-secret-key")

logger = logging.getLogger("mbt")
logging.basicConfig(level=logging.INFO)


@app.route("/")
def home():
    user = session.get("user")
    password = session.get("password")
    if user and password:
        logger.info("Rendering index.html for user")
        return render_template("index.html", user=user)
    logger.info("Rendering login.html (no user in session)")
    return render_template("login.html")


if __name__ == "__main__":
    # Allow overriding the port via PORT environment variable and
    # fall back to a secondary port if the preferred one is in use.
    preferred = int(os.environ.get("PORT", 5050))
    fallback = int(os.environ.get("FALLBACK_PORT", 5000))
    try:
        app.run(debug=True, port=preferred)
    except OSError as e:
        logger.warning(f"Port {preferred} unavailable: {e}; trying fallback {fallback}")
        app.run(debug=True, port=fallback)
