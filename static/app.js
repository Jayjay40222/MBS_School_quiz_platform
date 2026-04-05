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
function show(id, el, push = true) {
    document.querySelectorAll(".page").forEach(p => p.classList.remove("active"));
    document.querySelectorAll(".tab").forEach(t => t.classList.remove("active"));

    document.getElementById(id).classList.add("active");
    if (el) el.classList.add("active");

    if (push) history.pushState({ page: id }, "", "/" + id);
}

window.addEventListener("popstate", (event) => {
    const id = event.state?.page || "home";
    const tab = document.querySelector(`.tab[onclick*="show('${id}'`);
    if (tab) show(id, tab, false);
});

// Initialize UI on page load
document.addEventListener("DOMContentLoaded", initUI);