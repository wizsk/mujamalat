const ttl = document.querySelector("#title");
const word = ttl ? ttl.dataset.title : "";
const selectCustomAfterDailouge = document.getElementById(
  "selectCustomAfterDailouge",
);
const customAfterSelect = document.getElementById(
  "selectCustomAfterDailougeSelect",
);
const customAfterInput = document.getElementById(
  "selectCustomAfterDailougeInput",
);

const ord = document.getElementById("ord");
const ordLSN = () => `${window.location.pathname}-ord`;

document.addEventListener("DOMContentLoaded", () => {
  const rwh = (url) => window.history.replaceState(null, "", url);

  const curSearch = new URLSearchParams(window.location.search);

  const v = window.localStorage.getItem(ordLSN());
  if (!v) {
    if (curSearch.get("ord"))
      window.location.replace(window.location.pathname);

    ord.value = ord.options[0].value;
    rwh(`${window.location.pathname}`);
  } else {
    if (curSearch.get("ord") != v)
      window.location.replace(`${window.location.pathname}?ord=${v}`);

    ord.value = v;
    rwh(`${window.location.pathname}?ord=${v}`);
  }
});

ord.onchange = () => {
  const v = ord.value;
  if (v == "old") {
    window.localStorage.removeItem(ordLSN());
    window.location.href = window.location.pathname;
  } else {
    window.localStorage.setItem(ordLSN(), v);
    window.location.href = `${window.location.pathname}?ord=${v}`;
  }
};

function showAfter(d) {
  let days;
  if (d < 0) days = "r";
  else days = d;
  postThen(`${window.location.pathname}?w=${word}&after=${days}`, () =>
    window.location.reload(),
  );
}

customAfterInput.addEventListener("keypress", (e) => {
  if (e.key === "Enter") customAfter();
});

customAfterSelect.addEventListener(
  "change",
  () => (customAfterInput.value = ""),
);

function customAfter() {
  let val = customAfterInput.value.trim();
  val = val ? val : customAfterSelect.value;
  const days = parseInt(val);

  if (days < 0 || isNaN(days) || !/^[0-9]+$/.test(val)) {
    alert(`Error: Invalid number of days provided: '${val}'`);
    return;
  }
  selectCustomAfterDailouge.close();
  showAfter(days);
}

function dontShow() {
  if (confirm(`هل تريد إخفاء: ${word}`))
    postThen(`${window.location.pathname}?w=${word}&dont_show=true`, () =>
      window.location.reload(),
    );
}

function postThen(url, func) {
  fetch(url, { method: "POST" })
    .then((res) => {
      if (res.ok) {
        if (func) func();
      }
    })
    .catch((err) => {
      console.error(err);
      alert("Something went wrong (check console)");
    });
}
