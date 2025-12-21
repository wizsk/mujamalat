const histS = document.getElementById("hist-s");

histS.oninput = () => {
  const itms = document
    .querySelector("#hist-items")
    .querySelectorAll(".hist-item-div");
  const v = histS.value.trim();

  for (let i of itms) {
    if (v == "") i.classList.remove("hidden");

    const ar = i.dataset.ar;
    if (!ar || !ar.includes(v)) {
      i.classList.add("hidden");
    } else {
      i.classList.remove("hidden");
    }
  }
};
