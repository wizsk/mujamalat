const dialog = document.getElementById("dialog");
const isRequesting = {
    val: false,
    set: (nval) => {
        this.val = nval;
        dfsave.disabled = nval;
        dfres.disabled = nval;
        dh.disabled = nval;
    },
    show: () => {
        if (this.val) {
            alert('already requesting wait a bit');
        }
    }
}
let isShowingModal = false;

document.querySelectorAll(".row").forEach((r) => {
    r.addEventListener('click', () => {
        showModal(r);
    })
})

const dw = document.getElementById("dw");
/** future div */
const dfd = document.getElementById("dfd");
const df = document.getElementById("df");
// const dp = document.getElementById("dp");
const dh = document.getElementById("dh");

const dfres = document.getElementById("dfreset"); // d future reset
const dfs = document.getElementById("dfs"); // d future reset
const dfi = document.getElementById("dfi"); // d future reset
const dfsave = document.getElementById("dfsave"); // d future

dfs.addEventListener('change', () => dfi.value = "");
dfi.addEventListener('keydown', (e) => {
    if (e.key === "Enter" && dfi.value
        && !isRequesting.val && isShowingModal)
        dfsave.click()
});

dfi.addEventListener('input', (e) => {
    const val = e.data;
    if (val && !/^[0-9]+$/.test(val)) {
        dfi.value = val.replaceAll(val, '');
        return;
    }

    const num = parseInt(dfi.value);
    dfi.value = num > 30 ? "30" : num < 1 ? "1" : dfi.value;
})

dialog.addEventListener('close', () => isShowingModal = false);

function showModal(r) {
    // reseting
    isRequesting.set(false);
    isShowingModal = true;

    const wordElm = r.querySelector("[data-w]");
    const word = wordElm.dataset.w;

    const future = r.querySelector("[data-f]");
    const past = r.querySelector("[data-p]");
    const dontShow = r.querySelector("[data-h]");
    const sleepFor = 2000;

    // console.log(word, future, past, dontShow);

    dw.innerText = word;
    df.innerText = future.dataset.f ? future.dataset.f : "Not set";

    let hidden = dontShow.dataset.h == "true";
    if (hidden) dfd.classList.add('hidden');
    dh.innerText = hidden ? "Show" : "Hide";

    dfsave.onclick = async () => {
        if (isRequesting.val) {
            isRequesting.show();
            return;
        }

        isRequesting.set(true);
        dfsave.innerText = "Saving…";

        let finalVal = ""
        if (dfi.value) {
            finalVal = dfi.value;
        } else if (dfs.value) {
            finalVal = dfs.value;
        } else {
            alert("Please select or enter ammount of days");
        }

        // save req

        let res;
        try {
            res = await post202(`${window.location.pathname}?w=${word}&after=${finalVal}&api=true`, true);
        } catch (err) {
            console.error(err)
        }

        if (res && res.ok) {
            df.innerText = res.data.f;
            future.dataset.f = res.data.f;
            future.innerText = res.data.f;
            future.dataset.fu = res.data.fu;
            past.dataset.pu = res.data.pu;
            past.innerText = res.data.p;
            past.dataset.pu = res.data.pu;
        } else {
            alert("Failed to set new date");
        }

        dfsave.innerText = "Save";
        isRequesting.set(false);
    }


    dfres.onclick = async () => {
        if (isRequesting.val) {
            isRequesting.show();
            return;
        }
        isRequesting.set(true);
        dfres.innerText = "Resetting…";

        // future reset req
        let success;
        try {
            success = await post202(`${window.location.pathname}?w=${word}&after=reset`);
        } catch (err) {
            console.error(err)
        }

        dfres.innerText = "Reset";
        if (success == true) {
            future.dataset.f = "";
            future.dataset.fu = "0";
            future.innerText = ""
            past.dataset.p = "";
            past.dataset.pu = "0";
            past.innerText = ""
            df.innerText = "Not set";
        }
        isRequesting.set(false);
    }

    dh.onclick = async () => {
        if (isRequesting.val) {
            isRequesting.show();
            return;
        }
        isRequesting.set(true);

        // optimistic
        hidden = !hidden;
        dh.innerText = hidden ? "Showing…" : "Hiding…";

        // req here
        try {
            if (true == !await post202(`${window.location.pathname}?w=${word}&dont_show=${hidden}`)) {
                hidden = !hidden;
            }
        } catch (err) {
            hidden = !hidden;
            console.error(err)
        }

        dh.innerText = "Reset";

        if (hidden) dfd.classList.add('hidden');
        else dfd.classList.remove('hidden');

        isRequesting.set(false);
        dh.innerText = hidden ? "Show" : "Hide";
        dontShow.dataset.h = hidden ? "true" : "false";
        dontShow.innerText = hidden ? "Hidden" : "Shown";
    }

    dialog.showModal();
}


async function post202(url, resJSON) {
    const r = await fetch(url, { method: "POST" });
    if (r && r.status == 202) {
        if (resJSON) {
            return { ok: true, data: await r.json() }
        }
        return true;
    }
    if (resJSON) {
        return null;
    }
    return false;
}