// key.js

document.addEventListener('keydown', (e) => {
    // console.log(e.code);
    if (e.ctrlKey) {
        switch (e.code) {
            case "KeyP":
                e.preventDefault();
                toggleChangeDict();
                break;

            case "KeyH":
            case "ArrowDown":
                e.preventDefault();
                querySelectorNextPre(true, 'sw-dict-item', 'sw-dict-item-selected');
                break;

            case "Semicolon":
            case "ArrowUp":
                e.preventDefault();
                querySelectorNextPre(false, 'sw-dict-item', 'sw-dict-item-selected');
                break;

            case "KeyK":
                e.preventDefault();
                querySelectorNextPre(true, 'querySelector-item', 'querySelector-item-selected');
                break;

            case "KeyJ":
                e.preventDefault();
                querySelectorNextPre(false, 'querySelector-item', 'querySelector-item-selected');
                break;
        }
        return;
    }


    if (e.code === "Escape") {
        if (document.activeElement === w) w.blur();
        else if (isChangeDictShwoing) toggleChangeDict();
        return;
    }

    if (isChangeDictShwoing) {
        if (document.activeElement !== changeDictInpt) changeDictInpt.focus();

        if (e.key !== "Enter" || (e.code && e.code !== "Enter")) return;

        if (changeDictInpt.value.trim() === "") {
            toggleChangeDict();
        }

        if (!selectDict(changeDictInpt.value, true)) {
            changeDictInpt.value = "";
            changeDictInpt.focus();
        }
        return;
    }

    if (document.activeElement === w) return;

    if (e.code && /^Digit[1-9]$/.test(e.code)) {
        const i = parseInt(e.code.at(5)) - 1;

        if (e.shiftKey && i < dicts.length) {
            dicts[i].click();
            return
        }

        const q = document.getElementsByClassName("querySelector-item");
        if (i < q.length) q[i].click();

        return;
    }

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

        case "KeyP":
            e.preventDefault();
            toggleChangeDict();
            break;

        case "KeyK":
        case "ArrowLeft":
            e.preventDefault();
            querySelectorNextPre(true, 'querySelector-item', 'querySelector-item-selected');
            break;

        case "KeyJ":
        case "ArrowRight":
            e.preventDefault();
            querySelectorNextPre(false, 'querySelector-item', 'querySelector-item-selected');
            break;

        case "KeyH":
        case "ArrowDown":
            e.preventDefault();
            querySelectorNextPre(true, 'sw-dict-item', 'sw-dict-item-selected');
            break;

        case "Semicolon":
        case "ArrowUp":
            e.preventDefault();
            querySelectorNextPre(false, 'sw-dict-item', 'sw-dict-item-selected');
            break;

    }

})


/** 
 * @param {boolean} next if true then goes to the next query otherwise preveious 
 * @param {string}  className classname for the buttons
 * @param {string}  id is the button wich is selected for now
 */
function querySelectorNextPre(next, className, id) {
    /** @type {NodeListOf<HTMLElement>} */
    const children = document.getElementsByClassName(className);
    if (children.length < 2) return;

    let curr = -1;
    for (let i = 0; i < children.length; i++) {
        if (children[i].id === id) {
            curr = i;
            break;
        }
    }

    if (curr < 0) return;

    if (next)
        curr = curr + 1 === children.length ? 0 : curr + 1;
    else
        curr = curr - 1 < 0 ? children.length - 1 : curr - 1;

    children[curr].click();
}
