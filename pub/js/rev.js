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
const rand = document.getElementById("rand");

document.addEventListener("DOMContentLoaded", () => {
  if (window.location.href.includes("rand=true")) {
    rand.checked = true;
  }
});

rand.onchange = () => {
  if (rand.checked) {
    window.history.replaceState(
      null,
      "",
      `${window.location.pathname}?rand=true`,
    );
  } else {
    window.history.replaceState(null, "", `${window.location.pathname}`);
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
