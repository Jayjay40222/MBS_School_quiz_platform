// Navigate to a subject quiz
function openSub(sub) {
    // localStorage.setItem("subject", sub);
    location.href = "/quiz/play";
}

// Save profile photo and class/name edits
function save() {
    const newName = document.getElementById("newName").value.trim();
    const newClass = document.getElementById("newClass").value.trim();
    const user = JSON.parse(localStorage.getItem("user") || "{}");

    if (newName) user.n = newName;
    if (newClass) user.cls = newClass;

    localStorage.setItem("user", JSON.stringify(user));
    location.reload();
}

function loadPhoto(event) {
    const file = event.target.files[0];
    const reader = new FileReader();

    reader.onload = function () {
        document.getElementById("profilePic").src = reader.result;
        const user = JSON.parse(localStorage.getItem("user") || "{}");
        user.photo = reader.result;
        localStorage.setItem("user", JSON.stringify(user));
    };

    reader.readAsDataURL(file);
}

// Tab navigation
// app.js
function setActiveTab(tabElement) {
    // Remove active from all tabs
    document.querySelectorAll(".nav .tab").forEach(t => {
        t.classList.remove("active");
        t.removeAttribute('aria-current');
    });
    // Add active to the clicked one
    tabElement.classList.add("active");
    tabElement.setAttribute('aria-current', 'page');
}

// Example for single-page tabs
document.querySelectorAll(".nav .tab").forEach(tab => {
    tab.addEventListener("click", function(e) {
        // set active immediately for snappy UI
        setActiveTab(this);
        // let the link behave normally (navigate) so server-side routing still works
    });
});

// Initialize active tab from current path
function initActiveFromPath() {
    let path = window.location.pathname || '/dashboard';
    // treat root '/' as '/dashboard' for matching
        // Normalize path: remove trailing slash (but keep single '/')
        path = path.replace(/\/+$/, '') || '/';
        // treat root '/' and '/dashboard' as '/home' for matching
        if (path === '/' || path === '/dashboard') path = '/home';
    // Find the tab with matching data-path or fallback to first tab
    const match = document.querySelector(`.nav .tab[data-path="${path}"]`);
    if (match) {
        setActiveTab(match);
    } else {
        // try matching by prefix (e.g. /quiz/play)
        const partial = document.querySelector(`.nav .tab[data-path][data-path="/quiz"]`);
        if (path.startsWith('/quiz') && partial) setActiveTab(partial);
    }
}

// Call on load
document.addEventListener('DOMContentLoaded', initActiveFromPath);