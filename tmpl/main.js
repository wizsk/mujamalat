// don't remve this comment
let selectedDict = "{{.Curr}}";
let selectedDictAr = "{{index .DictsMap .Curr}}";

let searhInvId;
const contentHolder = document.getElementById("content");
const dicts = document.getElementsByClassName('sw-dict-item');
const urlParams = new URLSearchParams(window.location.search);


document.addEventListener("DOMContentLoaded", () => {
    history.replaceState(
        { html: contentHolder.innerHTML, query: "{{.Query}}", title: document.title },
        "", `${window.location.pathname}?${urlParams.toString()}`);

    const selected = document.getElementById('sw-dict-item-selected');
    if (selected && selected.scrollIntoView) {
        selected.scrollIntoView({
            behavior: 'smooth',    // auto Or 'smooth' if you want animation
            block: 'nearest',
            inline: 'center'     // Center the element horizontally
        });
    }

    setSavedFontSize();

    if (w.value.length === 0)
        w.focus();
    // w.setSelectionRange(w.value.length, w.value.length);
});

w.oninput = () => {
    clearInterval(searhInvId);
    searhInvId = setTimeout(() => {
        const word = w.value.trim();
        if (word === "") return;

        for (let i = 0; i < dicts.length; i++) {
            const n = dicts[i].getAttribute('data-dict-name');
            dicts[i].href = `/${n}?w=${word}`;
        }

        console.log(`/content?dict=${selectedDict}&w=${w.value.trim()}`);

        urlParams.set('w', word);
        const newUrl = `${window.location.pathname}?${urlParams.toString()}`;
        document.title = word;

        fetch(`/content?dict=${selectedDict}&w=${word}`).then(async (r) => {
            if (r.ok) {
                const h = await r.text();
                contentHolder.innerHTML = h;
                document.title = `${selectedDictAr}: ${word}`;
                window.history.pushState({ html: h, query: word, title: `${selectedDictAr}: ${word}` },
                    '', newUrl);
            }
        }).catch((err) => {
            contentHolder.innerHTML =
                `<div style="text-align: center; margin-top: 4rem; color: var(--alert);">
                Cound't fetch results. Is the server running? Refresshing in 3seconds...
            </div>`;
            console.error(err);
            setTimeout(() => window.location.href = newUrl, 3000);
        })


    }, 100);
}

window.addEventListener("popstate", (e) => {
    if (e.state) {
        contentHolder.innerHTML = e.state.html;
        document.title = e.state.title;
        w.value = e.state.query;
    }
});

document.addEventListener('keydown', (e) => {
    // no composite key
    if (e.ctrlKey) {
        return;
    }

    if (e.code === "Escape") {
        if (document.activeElement === w)
            w.blur();
        return;
    }

    if (document.activeElement === w) return;


    const input = w;
    switch (e.code) {
        case "KeyS":
            e.preventDefault();
            input.focus();
            input.setSelectionRange(input.value.length, input.value.length);
            break;
        case "KeyI":
            e.preventDefault();
            input.focus();
            input.select();
            break;

        case "Equal":
            e.preventDefault();
            fontSizeInc();
            break;
        case "Minus":
            e.preventDefault();
            fontSizeDec();
            break;
        case "Digit0":
            e.preventDefault();
            resetFontSize();
            break;

        case "KeyU":
            e.preventDefault();
            scroolToTop();
            break;


        case "KeyT":
            e.preventDefault();
            changeColor();
            break;

    }

})


const fontDiffCVal = 2.0;

function saveFont(s) {
    window.localStorage.setItem("font-size-for-kamusssss", s)
}

function setSavedFontSize() {
    const s = window.localStorage.getItem("font-size-for-kamusssss");
    if (s) {
        document.body.style.fontSize = s;
        resetFont.classList.remove("hidden");
    }
}

function fontSizeInc() {
    const v = parseFloat(window.getComputedStyle(document.body, null).getPropertyValue("font-size"))
    const s = `${v + fontDiffCVal}px`;
    document.body.style.fontSize = s;
    saveFont(s);
    resetFont.classList.remove("hidden");
}

function fontSizeDec() {
    const v = parseFloat(window.getComputedStyle(document.body, null).getPropertyValue("font-size"))
    const s = `${v - fontDiffCVal}px`;
    document.body.style.fontSize = s;
    saveFont(s);
    resetFont.classList.remove("hidden");
}

function scroolToTop() { window.scrollTo({ top: 0, behavior: 'smooth' }); }

function resetFontSize() {
    console.log("fontsize reset-edd")
    localStorage.removeItem("font-size-for-kamusssss");
    document.body.style.fontSize = "";
    resetFont.classList.add("hidden");
}

plus.onclick = fontSizeInc;
minus.onclick = fontSizeDec;
up.onclick = scroolToTop;

resetFont.onclick = resetFontSize;