from flask import Flask

def create_app():
    app = Flask(__name__)
    
    # Import and register blueprints
    from .routes import auth_bp
    from .routes import leaderboard_bp
    from .routes import quiz_bp
    
    app.register_blueprint(auth_bp)
    app.register_blueprint(leaderboard_bp)
    app.register_blueprint(quiz_bp)

    return app