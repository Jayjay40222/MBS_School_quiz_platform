from flask import Blueprint, jsonify

quiz_bp = Blueprint("quiz", __name__)

@quiz_bp.route("/questions", methods=["GET"])
def get_questions():
    return jsonify({
        "questions": [
            {"q": "2+2?", "a": 4}
        ]
    })