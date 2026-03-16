(function () {
    window.__APP_LOADING_START__ = Date.now();

    var progressBar = document.getElementById("progressBar");
    var progressText = document.getElementById("progressText");
    if (!progressBar || !progressText) {
        return;
    }

    var progress = 0;
    var timer = setInterval(function () {
        var increment = Math.floor(Math.max(1, 10 * (1 - progress / 100)));
        progress = Math.min(99, progress + increment);
        progress = Math.floor(progress);

        progressBar.style.width = progress + "%";
        progressText.textContent = progress + "% Complete";

        if (progress >= 99) {
            clearInterval(timer);
        }
    }, 200);
})();
