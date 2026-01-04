const form = document.getElementById("form");
const tmpMode = document.getElementById("tmpMode");
const txt = document.getElementById("txt");
const hist = document.getElementById("hist");
const save = document.getElementById("save");
const histTitle = document.getElementById("hist-title");
const histItems = document.getElementById("hist-items");
const histCS = document.getElementById("hist-count");
const histSortType = document.getElementById("hist-sort-type");
let histItemsRev = false;
let histC = parseInt("{{len .}}");

form.onsubmit = (e) => {
  e.preventDefault();
  const val = txt.value.trim();
  if (val === "") return;

  const url = tmpMode.checked ? "/rd/tmp/" : "/rd/";

  fetch(url, {
    method: "POST",
    headers: { "Content-Type": "text/plain" },
    body: val,
  })
    .then(async (res) => {
      if (res.redirected) {
        window.location.href = res.url;
      } else {
        alert(`Something went wrong: ${await res.text()}`);
      }
    })
    .catch((e) => console.error(e));
};

tmpMode.addEventListener("click", () => {
  if (hist) {
    if (tmpMode.checked) {
      hist.classList.add("hidden");
    } else {
      hist.classList.remove("hidden");
    }
  }
});

const histSortLSK = "hist-sort-items-key";
const histTitleNorm = " (من جديد إلى قديم)";
const histTitleRev = " (من قديم إلى جديد)";
document.addEventListener("DOMContentLoaded", (e) => {
  tmpMode.checked = false;
  if (window.localStorage.getItem("dark"))
    document.documentElement.classList.add("dark");

  if (histTitle) {
    const val = window.localStorage.getItem(histSortLSK);
    if (val) histItemsRev = true;
    histItmsRev(histItemsRev, false);
  }
});

if (histTitle) {
  histTitle.onclick = () => {
    histItemsRev = !histItemsRev;
    histItmsRev(histItemsRev, true);
  };
}

function histItmsRev(rev, save) {
  if (rev) {
    sortButtons(true);
    histSortType.innerText = histTitleRev;
    if (save) window.localStorage.setItem(histSortLSK, "rev");
  } else {
    sortButtons();
    histSortType.innerText = histTitleNorm;
    if (save) window.localStorage.removeItem(histSortLSK);
  }
}

const del = document.getElementsByClassName("del");
for (let i = 0; i < del.length; i++) {
  const cd = del[i];
  cd.onclick = (e) => {
    const pName =
      cd.dataset.name.slice(0, 50) + (cd.dataset.name.length > 50 ? "..." : "");

    console.log(pName);
    const c = confirm("هل تريد أن تمسح: " + pName + "؟");
    if (!c) return;
    fetch(cd.dataset.link, { method: "POST" })
      .then(async (r) => {
        if (r.ok && r.status === 202) {
          histCS.innerText = `${--histC}`.replace(/[0-9]/g, (d) =>
            String.fromCharCode(0x0660 + parseInt(d)),
          );
          histItems.removeChild(cd.parentElement);
          if (histItems.childElementCount === 0) {
            hist.style.display = "none";
          }
        } else {
          console.log(r);
          alert("Coun't delete: " + pName);
        }
      })
      .catch((err) => {
        alert("Coun't delete: " + pName);
        console.error(err);
      });
  };
}

// Select the container element

// Select all buttons (or rm-btn elements) inside the container

// Add click event listeners to all buttons
const pins = document.querySelectorAll(".pin");
pins.forEach((p) => {
  p.addEventListener("click", async () => {
    const npinned = p.parentElement.dataset.pin == "0";

    const pName =
      p.dataset.name.slice(0, 50) + (p.dataset.name.length > 50 ? "..." : "");

    if (!npinned && !confirm("هل ترغب في إزالة التثبيت: " + pName + "؟"))
      return;

    const u = `/rd/entryEdit?sha=${p.dataset.sha}&pin=${npinned ? "true" : "false"}`;
    let res = null;
    try {
      res = await fetch(u, { method: "POST" });
    } catch (err) {
      console.error(err);
    }

    if (!res || !res.ok || res.status !== 202) {
      alert("لم يتمكن من التثبيت/إلغاء التثبيت: " + pName);
      return;
    }

    if (npinned) {
      p.classList.add("pinned");
      p.parentElement.dataset.pin = "1";
    } else {
      p.classList.remove("pinned");
      p.parentElement.dataset.pin = "0";
    }
    sortButtons(histItemsRev);
  });
});

function sortButtons(rev) {
  const container = document.querySelector("#hist-items");
  const buttons = Array.from(container.querySelectorAll(".hist-item-div"));

  const pinnedHistItems = buttons.filter(
    (button) => button.getAttribute("data-pin") === "1",
  );
  const nonPinnedHistItems = buttons.filter(
    (button) => button.getAttribute("data-pin") !== "1",
  );

  pinnedHistItems.sort((a, b) => {
    if (rev)
      return (
        parseInt(b.getAttribute("data-idx")) -
        parseInt(a.getAttribute("data-idx"))
      );

    return (
      parseInt(a.getAttribute("data-idx")) -
      parseInt(b.getAttribute("data-idx"))
    );
  });

  nonPinnedHistItems.sort((a, b) => {
    if (rev)
      return (
        parseInt(b.getAttribute("data-idx")) -
        parseInt(a.getAttribute("data-idx"))
      );

    return (
      parseInt(a.getAttribute("data-idx")) -
      parseInt(b.getAttribute("data-idx"))
    );
  });

  const sortedButtons = [...pinnedHistItems, ...nonPinnedHistItems];

  container.innerHTML = "";
  sortedButtons.forEach((button) => container.appendChild(button));
}

document.getElementById("tgl-cursor").onclick = () => {
  const text = ta.value;
  if (text.length == 0) return;
  const cursorPos = ta.selectionStart;
  const mid = text.length / 2;

  if (cursorPos > mid) {
    // go to top
    ta.selectionStart = ta.selectionEnd = 0;
  } else {
    // go to bottom
    const end = text.length;
    ta.selectionStart = ta.selectionEnd = end;
  }

  ta.focus();
};

document.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.shiftKey) && e.code == "Enter") {
    e.preventDefault();
    form.submit.click();
  }

  if (document.activeElement === txt) {
    if (e.code === "Escape") txt.blur();
    return;
  }

  if (!e.ctrlKey) return;

  const input = ta;
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

    case "KeyR":
      e.preventDefault();
      goToReader.click();
      break;
  }
});
