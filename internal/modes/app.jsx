import { useEffect, useState } from "react";
import { GetApps, BlockApps, UnblockApps } from "../wailsjs/go/main/AppManager";

function App() {
  const [apps, setApps] = useState([]);
  const [selected, setSelected] = useState([]);

  useEffect(() => {
    GetApps().then(setApps);
  }, []);

  const toggleSelect = (app) => {
    setSelected((prev) =>
      prev.includes(app.exec)
        ? prev.filter((a) => a !== app.exec)
        : [...prev, app.exec]
    );
  };

  return (
    <div style={{ padding: "20px" }}>
      <h2>Select Apps to Block</h2>

      <div style={{ maxHeight: "400px", overflow: "auto" }}>
        {apps.map((app, index) => (
          <div key={index}>
            <label>
              <input
                type="checkbox"
                onChange={() => toggleSelect(app)}
              />
              {app.name}
            </label>
          </div>
        ))}
      </div>

      <button onClick={() => BlockApps(selected)}>
        Start Focus Mode
      </button>

      <button onClick={() => UnblockApps(selected)}>
        Stop Focus Mode
      </button>
    </div>
  );
}

export default App;