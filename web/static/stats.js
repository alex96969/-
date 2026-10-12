"use strict";

(() => {
  const rubles = new Intl.NumberFormat("ru-RU", {style: "currency", currency: "RUB"});
  const palette = ["#305f9d", "#f5bb58", "#74a6c0", "#9ca1cd", "#d98f78", "#7aa793", "#bcc5d4"];

  function buildTable(rows) {
    const table = document.createElement("table");
    const header = table.createTHead().insertRow();
    for (const name of ["Категория / месяц", "Расходы", "Записей"]) {
      const cell = document.createElement("th");
      cell.scope = "col";
      cell.textContent = name;
      header.append(cell);
    }
    const body = table.createTBody();
    for (const item of rows) {
      const row = body.insertRow();
      for (const value of [item.label, rubles.format(item.amount / 100), item.count]) {
        row.insertCell().textContent = value;
      }
    }
    return table;
  }

  class StatisticsPanel {
    constructor(kind) {
      this.kind = kind;
      this.form = document.getElementById(`${kind}-filters`);
      this.message = document.getElementById(`${kind}-message`);
      this.table = document.getElementById(`${kind}-table`);
      this.canvas = document.getElementById(`${kind}-chart`);
      this.chart = null;
      this.request = null;
      this.form.addEventListener("submit", event => {
        event.preventDefault();
        this.refresh();
      });
      this.refresh();
    }

    async refresh() {
      this.request?.abort();
      const request = new AbortController();
      this.request = request;
      this.chart?.destroy();
      this.chart = null;
      this.table.replaceChildren();
      this.message.textContent = "Загружаем данные…";
      try {
        const query = new URLSearchParams(new FormData(this.form));
        const response = await fetch(`/api/stats/by-${this.kind}?${query}`, {
          credentials: "same-origin", signal: request.signal
        });
        const rows = await response.json();
        if (!response.ok) throw new Error(rows.error || "Не удалось получить данные.");
        if (request.signal.aborted) return;
        this.table.replaceChildren(buildTable(rows));
        const total = rows.reduce((sum, row) => sum + row.amount, 0);
        this.message.textContent = total === 0 ? "В этом периоде расходов нет." : `Итого: ${rubles.format(total / 100)}`;
        if (typeof Chart === "undefined") {
          this.message.textContent += " График недоступен. Все значения показаны ниже.";
          this.table.closest("details").open = true;
          return;
        }
        if (total === 0 && this.kind === "category") return;
        this.draw(rows);
      } catch (error) {
        if (error.name !== "AbortError") this.message.textContent = error.message;
      }
    }

    draw(rows) {
      const categories = this.kind === "category";
      const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {display: categories, position: "bottom"},
          tooltip: {callbacks: {label: context => rubles.format(categories ? context.parsed : context.parsed.y)}}
        }
      };
      if (!categories) options.scales = {y: {beginAtZero: true}};
      this.chart = new Chart(this.canvas, {
        type: categories ? "pie" : "bar",
        data: {
          labels: rows.map(row => row.label),
          datasets: [{
            label: "Расходы, ₽",
            data: rows.map(row => row.amount / 100),
            backgroundColor: categories ? palette : "#f5bb58",
            borderColor: "#fff", borderWidth: categories ? 2 : 0,
            borderRadius: categories ? 0 : 5
          }]
        },
        options
      });
    }
  }

  for (const kind of ["category", "month"]) new StatisticsPanel(kind);
})();
