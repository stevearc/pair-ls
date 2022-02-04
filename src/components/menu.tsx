import * as React from "react";
import MenuIcon from "@mui/icons-material/Menu";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import IconButton from "@mui/material/IconButton";
import FormGroup from "@mui/material/FormGroup";
import FormControlLabel from "@mui/material/FormControlLabel";
import Switch from "@mui/material/Switch";
import { AppContext } from "../state";
import ColorChooser from "./color_chooser";
const { useContext, useState } = React;

export default function MenuComponent() {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [colorChooserOpen, setColorChooserOpen] = React.useState(false);
  const { state, dispatch } = useContext(AppContext);
  const open = Boolean(anchorEl);
  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <div>
      <IconButton
        id="menu-button"
        size="large"
        edge="start"
        color="inherit"
        aria-label="menu"
        aria-controls={open ? "basic-menu" : undefined}
        aria-haspopup="true"
        aria-expanded={open ? "true" : undefined}
        sx={{ marginLeft: "4px", marginRight: 0 }}
        onClick={handleClick}
      >
        <MenuIcon />
      </IconButton>
      <Menu
        id="basic-menu"
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        MenuListProps={{
          "aria-labelledby": "menu-button",
        }}
      >
        <MenuItem
          onClick={() => {
            dispatch({ type: "toggleFollow" });
          }}
        >
          <FormGroup>
            <FormControlLabel
              control={<Switch color="secondary" checked={state.follow} />}
              label={state.follow ? "Following" : "Follow"}
            />
          </FormGroup>
        </MenuItem>
        <MenuItem
          onClick={() => {
            setColorChooserOpen(true);
            handleClose();
          }}
        >
          Color Scheme
        </MenuItem>
      </Menu>
      <ColorChooser
        colorscheme={state.colorscheme}
        open={colorChooserOpen}
        onChange={(colorscheme) => {
          dispatch({
            type: "setColorScheme",
            colorscheme,
          });
        }}
        onClose={() => setColorChooserOpen(false)}
      />
    </div>
  );
}
