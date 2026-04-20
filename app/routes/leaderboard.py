from flask import Blueprint, jsonify

leaderboard_bp = Blueprint("leaderboard", __name__)

@leaderboard_bp.route("/leaderboard", methods=["GET"])
def get_leaderboard():
    return jsonify({
        "leaderboard": [
            {"username": "Alice", "score": 100},
            {"username": "Bob", "score": 90}
        ]
    })
