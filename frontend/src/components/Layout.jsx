import Header from "./Header";
import { Outlet } from "react-router-dom";

const Layout = () => {
    return (
        <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
            <Header />
            <main className="main-content with-layout">
                <Outlet />
            </main>
        </div>
    )
}

export default Layout;
