// don't remove this comment
let selectedDict = "{{.Curr}}";
// if false then the dictionary is open or reading :D
let lastClikedWord = null;
let readerMode = "{{if .RDMode}}y{{end}}" === "y";
let selectedDictAr = "{{index .DictsMap .Curr}}";
const contentHolder = document.getElementById("content");
const navSpace = document.getElementById("navSpace");
const navSetHeightDelay = 100; // ms
const dicts = document.getElementsByClassName("sw-dict-item");
const changeDict = document.getElementById("change-dict");
const changeDictInpt = document.getElementById("change-dict-inpt");
const w = document.getElementById("w");
const wBtn = document.getElementById("w-btn");
const input = w;
let currWord = "{{if .Queries}}{{index .Queries .Idx}}{{end}}";
let preQuery = "{{.Query}}"; // this is current query belive it or not! lol
let queryIdx = 0;
try {
  queryIdx = parseInt("{{.Idx}}");
} catch (er) {
  console.error(er);
  queryIdx = 0;
}
let isChangeDictShwoing = false;
let scrollOnSearch = false;

const infoDialog = document.getElementById("infoDialog");

const dictHighContainer = document.getElementById("highWordContainer");
const dictHighWord = document.getElementById("highWord");
const dictHighHiBtn = document.getElementById("highWordHiBtn");
const dictHighNoteBtn = document.getElementById("highWordNoteBtn");

let resizeTimoutId;
window.addEventListener("resize", () => {
  clearInterval(resizeTimoutId);
  resizeTimoutId = setTimeout(() => {
    navSpace.style.height = `${nav.offsetHeight + 20}px`;
    showHideNav(true);
    // {{if .RDMode}}
    setPopUpPos();
    // {{end}}
  }, 100);
});

document.addEventListener("DOMContentLoaded", () => {
  window.localStorage.getItem("dark") &&
    document.documentElement.classList.add("dark");

  setSavedFontSize();

  setTimeout(() => {
    // it's a promise btw
    setNavHeight();
  }, navSetHeightDelay);

  if (window.localStorage.getItem(getScrollOnSearchLSN())) {
    scrollOnSearch = true;
  }

  // wBtnTgl(); // handeled by the go template

  // {{if not .RDMode}}
  if (scrollOnSearch) {
    setTimeout(() => {
      const target = document.querySelector(".search-hi");
      if (target) {
        target.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }, navSetHeightDelay + 20);
  }
  let selected = document.getElementById("sw-dict-item-selected");
  if (selected && selected.scrollIntoView) {
    selected.scrollIntoView({
      // behavior: 'smooth',    // auto Or 'smooth' if you want animation
      block: "nearest",
      inline: "center", // Center the element horizontally
    });
  }

  selected = document.getElementById("querySelector-item-selected");
  if (selected && selected.scrollIntoView) {
    selected.scrollIntoView({
      // behavior: 'smooth',    // auto Or 'smooth' if you want animation
      block: "nearest",
      inline: "center", // Center the element horizontally
    });
  }

  if (w.value.length === 0) w.focus();
  let p = "";
  if (currWord !== "") {
    p = `?w=${preQuery}&idx=${queryIdx}`;
  }
  window.history.replaceState(null, "", `${window.location.pathname}${p}`);
  // {{end}}
});

w.onfocus = () => showHideNav(true);

let lastScrollTop = 0;
let navHidden = false;
const threshold = 5; // adjust as needed
window.addEventListener(
  "scroll",
  function () {
    const currentScroll =
      window.pageYOffset || document.documentElement.scrollTop;

    // Only toggle if movement is greater than threshold
    if (Math.abs(currentScroll - lastScrollTop) > threshold) {
      showHideNav(currentScroll < lastScrollTop);
      lastScrollTop = currentScroll;
      // {{if .RDMode}}
      closePopup();
      // {{end}}
    }
  },
  { passive: true },
);

// {{if .RDMode}}
let lastScrollTopDict = 0;
dict_container.addEventListener(
  "scroll",
  function () {
    let currentScroll = dict_container.scrollTop;
    if (Math.abs(currentScroll - lastScrollTopDict) > threshold) {
      showHideNav(currentScroll < lastScrollTopDict);
      lastScrollTopDict = currentScroll;
    }
  },
  { passive: true },
);
// {{end}}

form.onsubmit = (e) => {
  e.preventDefault();
  w.blur();
  // {{if not .RDMode}}
  window.location.href = `${window.location.pathname}?w=${w.value}&idx=${queryIdx}`;
  // {{end}}
};

wBtn.onclick = () => {
  w.value = "";
  w.focus();
  if (w.oninput) w.oninput();
  wBtn.classList.add("hidden");
};

let searhInvId;
w.oninput = () => {
  wBtnTgl();
  clearInterval(searhInvId);

  searhInvId = setTimeout(async () => {
    const wv = w.value;
    // cleaning
    const queryArr = w.value.split(" ").filter((e) => e != "");
    const query = queryArr.join(" ");

    const curPos = w.selectionEnd;
    let curWord = null; // word under the cursor
    let wordIdx = -1;
    if (curPos == wv.length) {
      wordIdx = queryArr.length - 1;
      curWord = queryArr[wordIdx];
    } else {
      for (let i = 0; i < wv.length; ) {
        while (i < wv.length && wv[i] == " ") {
          i++;
        }
        if (i >= wv.length) break;

        wordIdx++;
        const next = wv.slice(i, wv.length);
        const nextSpace = next.indexOf(" ");
        // the end of and there are no spaces
        if (nextSpace < 0) {
          curWord = next;
          break;
        } else {
          const cw = next.slice(0, nextSpace);
          i += cw.length;
          while (i < wv.length && wv[i] == " ") {
            i++;
          }
          if (curPos < i) {
            curWord = cw;
            break;
          }
        }
      }
    }

    // console.log(curPos, curWord, wordIdx);
    // console.log(queryArr)

    const word = curWord;
    currWord = word;
    queryIdx = wordIdx;

    // set it as preQuery
    preQuery = wv;

    // its async hence non blocking u stupid
    getResAndShow(word);
    fetchAndSetDictHighWordState(word);

    querySelector.innerHTML = "";
    if (queryArr.length > 1) {
      let b = "";
      for (let i = 0; i < queryArr.length; i++) {
        const v = queryArr[i];

        b += `<button onclick='changeQueryIdx(this, ${JSON.stringify(v)}, ${i})'
                class="querySelector-item" id="${queryIdx == i ? "querySelector-item-selected" : ""}">
                ${v}</button>`;
      }
      querySelector.innerHTML = b;
      setTimeout(() => {
        const el = document.getElementById("querySelector-item-selected");
        if (el) el.scrollIntoView({
          behavior: 'smooth',    // auto Or 'smooth' if you want animation
          block: "nearest",
          inline: "center", // Center the element horizontally
        });
      }, 100);
      querySelector.classList.remove("hidden");
    } else {
      querySelector.innerHTML = "";
      querySelector.classList.add("hidden");
    }

    setTimeout(() => {
      setNavHeight();
    }, navSetHeightDelay);

    // {{if not .RDMode}}
    document.title = `${selectedDictAr}${word ? ": " + word : ""}`;
    const q = query ? `?w=${query}&idx=${queryIdx}` : "";
    window.history.replaceState(null, "", `${window.location.pathname}${q}`);
    // {{end}}
  }, 300);
};

async function getResAndShow(word) {
  if (!word || word === "") {
    contentHolder.innerHTML = `{{template "not-found"}}`;
    return;
  }

  contentHolder.innerHTML = `{{template "wait"}}`;

  // console.log(`req: /content?dict=${selectedDict}&w=${word}`);
  const r = await fetch(`/content?dict=${selectedDict}&w=${word}`).catch(
    (err) => console.error(err),
  );

  if (r && r.ok) {
    console.log(
      `cached:`,
      r.headers.get("X-From-Cache"),
      `/content?dict=${selectedDict}&w=${word}`,
    );
    const h = await r.text();
    contentHolder.innerHTML = h;
    if (scrollOnSearch) {
      const target = document.querySelector(".search-hi");
      if (target) {
        target.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }
  } else {
    contentHolder.innerHTML = `{{template "server-issue"}}`;
  }
}

function getFullHeight(element) {
  const rect = element.getBoundingClientRect();
  const style = window.getComputedStyle(element);

  const marginTop = parseFloat(style.marginTop);
  const marginBottom = parseFloat(style.marginBottom);
  const paddingTop = parseFloat(style.paddingTop);
  const paddingBottom = parseFloat(style.paddingBottom);
  const borderTop = parseFloat(style.borderTopWidth);
  const borderBottom = parseFloat(style.borderBottomWidth);

  // Full height = content height + padding + border + margin
  const fullHeight =
    rect.height +
    marginTop +
    marginBottom +
    paddingTop +
    paddingBottom +
    borderTop +
    borderBottom;
  return fullHeight;
}

function enToArNum(n) {
  return `${n}`.replace(/[0-9]/g, (d) =>
    String.fromCharCode(0x0660 + parseInt(d)),
  );
}

function changeColor(darkMode) {
  if (darkMode) {
    document.documentElement.classList.add("dark");
    window.localStorage.setItem("dark", "t");
  } else {
    document.documentElement.classList.remove("dark");
    window.localStorage.removeItem("dark");
  }

  try {
    dark.checked = darkMode ? true : false;
  } catch (e) {}
}

function wBtnTgl(focus) {
  if (w.value) wBtn.classList.remove("hidden");
  else wBtn.classList.add("hidden");
}

/** LSN = local storage name */
function getScrollOnSearchLSN() {
  // return `${window.location.pathname}-`
  return "scroll-on-search";
}

// called after high or rm high
let fetchAndSetDictHighWordStateCallBack;

// alhi alredy high
// alhi and callb is used in the reader page
async function fetchAndSetDictHighWordState(word, alhi, callb) {
  dictHighContainer.classList.add("hidden");

  let r;
  if (alhi == null) {
    try {
      r = await fetch(`/rd/high_has?w=${word}`, { method: "POST" });
    } catch (err) {
      console.error(err);
    }
  }

  if (alhi == null && !r) {
    console.log("what is the server running?");
    return;
  }

  const w = alhi != null ? word : r ? await r.text() : "";
  if (!w) return;

  const isHi = alhi != null ? alhi : r ? r.status == 202 : false;

  dictHighHiBtn.dataset.ishi = isHi ? "true" : "false";
  dictHighWord.textContent = w;
  dictHighHiBtn.dataset.oar = w;
  dictHighNoteBtn.dataset.oar = w;
  dictHighContainer.classList.remove("hidden");

  fetchAndSetDictHighWordStateCallBack = callb;
}

dictHighHiBtn.onclick = async () => {
  dictHighHiBtn.disabled = true;
  const isHi = dictHighHiBtn.dataset.ishi == "true";
  const word = dictHighHiBtn.dataset.oar;

  let ru = `/rd/high?w=${word}`;
  ru += isHi ? "&del=true" : "&add=true";

  let res;
  try {
    res = await fetch(ru, { method: "POST" });
  } catch (err) {
    console.error(err);
    dictHighHiBtn.disabled = false;
    return;
  }

  if (res && res.status == 202) {
    dictHighHiBtn.dataset.ishi = isHi ? "false" : "true";
    if (fetchAndSetDictHighWordStateCallBack)
      fetchAndSetDictHighWordStateCallBack(word, !isHi);
  }
  dictHighHiBtn.disabled = false;
};


function scrollToEl(el) {
  if (el) el.scrollIntoView({
    behavior: 'smooth',    // auto Or 'smooth' if you want animation
    block: "nearest",
    inline: "center", // Center the element horizontally
  });
}