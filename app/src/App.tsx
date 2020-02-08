import React, { useState, useEffect, useCallback, useRef } from 'react';
import clsx from 'clsx';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemSecondaryAction from '@material-ui/core/ListItemSecondaryAction';
import ListItemText from '@material-ui/core/ListItemText';
import Switch from '@material-ui/core/Switch';
import Button from '@material-ui/core/Button';
import Container from '@material-ui/core/Container';
import { withSnackbar, WithSnackbarProps, VariantType } from 'notistack';
import { library } from '@fortawesome/fontawesome-svg-core';
import { faTemperatureHigh, faChartBar, faTint, faFan, faWindowMaximize, faFire, faHdd } from '@fortawesome/free-solid-svg-icons';
import { faMemory, faLaptop, faSlidersH, faBug, faLayerGroup, faMicrochip, faBolt, faMoneyBillAlt } from '@fortawesome/free-solid-svg-icons';
import { faBurn, faPlug, faShower, faClock, faStopwatch, faToggleOn } from '@fortawesome/free-solid-svg-icons';
import { faKey } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Fab from '@material-ui/core/Fab';
import CheckIcon from '@material-ui/icons/Check';
import CloseIcon from '@material-ui/icons/Close';
import ExpandLess from '@material-ui/icons/ExpandLess';
import ExpandMore from '@material-ui/icons/ExpandMore';
import Collapse from '@material-ui/core/Collapse';
import Badge from '@material-ui/core/Badge';
import Websocket from 'react-websocket';
import * as qs from 'query-string';

import { useStyles } from './AppStyle';
import './App.css';

library.add(faTemperatureHigh, faKey);
library.add(faChartBar, faKey);
library.add(faTint, faKey);
library.add(faFan, faKey);
library.add(faWindowMaximize, faKey);
library.add(faFire, faKey);
library.add(faHdd, faKey);
library.add(faMemory, faKey);
library.add(faLaptop, faKey);
library.add(faSlidersH, faKey);
library.add(faBug, faKey);
library.add(faLayerGroup, faKey);
library.add(faMicrochip, faKey);
library.add(faBolt, faKey);
library.add(faMoneyBillAlt, faKey);
library.add(faBurn, faKey);
library.add(faShower, faKey);
library.add(faPlug, faKey);
library.add(faClock, faKey);
library.add(faStopwatch, faKey);
library.add(faToggleOn);

interface Props extends WithSnackbarProps { }

interface Row {
  Item: Item
  SubItems?: Array<Item>
}

interface Item {
  ID: string
  Type: string
  Label: string
  Value: string
  Img: string
  Unit: string
  LastUpdate: string
}

interface ItemRenderProps {
  className?: string
  baseUrl: string

  ID: string
  Type: string
  Label: string
  Value: string
  Img: string
  Unit: string
  SubItems?: Array<Item>
  LastUpdate: string
}

interface IconRenderProps {
  img: string
}

const RenderIcon: React.FC<IconRenderProps> = (props) => {
  switch (props.img) {
    case "temperature":
      return (<FontAwesomeIcon icon="temperature-high" />);
    case "humidity":
      return (<FontAwesomeIcon icon="tint" />);
    case "fan":
      return (<FontAwesomeIcon icon="fan" />);
    case "window":
      return (<FontAwesomeIcon icon="window-maximize" />);
    case "radiator":
      return (<FontAwesomeIcon icon="fire" />);
    case "hdd":
      return (<FontAwesomeIcon icon="hdd" />);
    case "mem":
      return (<FontAwesomeIcon icon="memory" />);
    case "sys":
      return (<FontAwesomeIcon icon="laptop" />);
    case "settings":
      return (<FontAwesomeIcon icon="sliders-h" />);
    case "dev":
      return (<FontAwesomeIcon icon="bug" />);
    case "group":
      return (<FontAwesomeIcon icon="layer-group" />);
    case "cpu":
      return (<FontAwesomeIcon icon="microchip" />);
    case "electricity":
      return (<FontAwesomeIcon icon="bolt" />);
    case "price":
      return (<FontAwesomeIcon icon="money-bill-alt" />);
    case "boiler":
      return (<FontAwesomeIcon icon="burn" />);
    case "shower":
      return (<FontAwesomeIcon icon="shower" />);
    case "plug":
      return (<FontAwesomeIcon icon="plug" />);
    case "clock":
      return (<FontAwesomeIcon icon="clock" />);
    case "timer":
      return (<FontAwesomeIcon icon="stopwatch" />);
    case "switch":
      return (<FontAwesomeIcon icon="toggle-on" />);
  }

  return (
    <FontAwesomeIcon icon="chart-bar" />
  )
};

const RenderSwitch: React.FC<ItemRenderProps> = (props) => {
  const [checked, setChecked] = React.useState(props.Value == "ON");

  const onChange = (event: any) => {
    setChecked(!checked)

    fetch(`${props.baseUrl}/item/${props.ID}`, {
      method: 'POST',
      body: !checked ? "ON" : "OFF"
    })
  }

  return (
    <Switch
      edge="end"
      color="primary"
      onChange={onChange}
      checked={props.Value == "ON"}
    />
  )
};

const RenderButton: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();

  return (
    <Button variant="contained" color="primary" className={props.Value === "ON" ? classes.btnGreen : ""}>
      {props.Label}
    </Button>
  )
};

const RenderValue: React.FC<ItemRenderProps> = (props) => {
  return (
    <Typography>{props.Value}{props.Unit}</Typography>
  )
};

interface groupRenderProps {
  items?: Array<Item>
  isOpen: boolean
}

const RenderGroupButton: React.FC<groupRenderProps> = (props) => {
  if (props.items) {
    if (!props.isOpen) {
      return (<ExpandLess />);
    }
    return (<ExpandMore />);
  }
  return (
    <svg viewBox="0 0 24 24" />
  )
};

const RenderState: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();

  if (props.Value === "ON") {
    return (
      <Fab color="primary" aria-label="OK" size="small" className={classes.fabGreen}>
        <CheckIcon />
      </Fab>
    )
  }
  return (
    <Fab color="primary" aria-label="KO" size="small" className={classes.fabRed}>
      <CloseIcon />
    </Fab>
  )
};

const RenderType: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();

  switch (props.Type) {
    case "switch":
      return (
        <RenderSwitch {...props} />
      )
    case "value":
      return (
        <RenderValue {...props} />
      )
    case "state":
      return (
        <RenderState {...props} />
      )
    case "button":
      return (
        <RenderButton {...props} />
      )
    case "timer":
      const value = props.Value === "OFF" ? "OFF" : "ON";
      const count = (props.Value !== "OFF" && props.Value !== "ON") ? props.Value : "0"

      return (
        <Badge color="primary" badgeContent={count}>
          <Button variant="contained" color="primary" className={props.Value === "ON" ? classes.btnGreen : classes.btnRed}>
            {value}
          </Button>
        </Badge>
      )
    default:
      return (
        <React.Fragment />
      )
  }
};

const RenderItem: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();
  const [open, setOpen] = React.useState(false);

  return (
    <React.Fragment>
      <ListItem className={props.className} onClick={() => { setOpen(!open) }}>
        <ListItemIcon>
          <React.Fragment>
            <RenderGroupButton items={props.SubItems} isOpen={open} />
            <RenderIcon img={props.Img} />
          </React.Fragment>
        </ListItemIcon>
        <ListItemText primary={props.Label} secondary={props.LastUpdate ? "Updated: " + props.LastUpdate : ""} />
        <ListItemSecondaryAction>
          <RenderType {...props} />
        </ListItemSecondaryAction>
      </ListItem>
      {props.SubItems &&
        <Collapse in={open} timeout="auto" unmountOnExit>
          <List component="div" disablePadding>
            {props.SubItems.map((item: Item) => (
              <RenderItem key={item.ID} className={classes.nested}
                ID={item.ID}
                Type={item.Type}
                Img={item.Img}
                Label={item.Label}
                Value={item.Value}
                Unit={item.Unit}
                LastUpdate={item.LastUpdate}
                baseUrl={props.baseUrl}
              />
            ))}
          </List>
        </Collapse>
      }
    </React.Fragment>
  )
};

const updateData = (item: Item, rows: Array<Row>) => {
  rows.forEach((row: Row, i: number) => {
    if (row.Item.ID === item.ID) {
      row.Item = item;
      return
    }

    if (row.SubItems) {
      row.SubItems.forEach((curr: Item, i: number) => {
        if (curr.ID === item.ID && row.SubItems) {
          row.SubItems[i] = item
        }
      });
    }
  });
};

const App: React.FC<Props> = (props) => {
  const classes = useStyles();
  const [data, setData] = useState({ Rows: Array<Row>() });
  const [intervalID, setIntervalID] = useState(0);
  const wsRef = useRef<any>(null);

  const address = qs.parse(window.location.search).address;
  const username = qs.parse(window.location.search).username;
  const password = qs.parse(window.location.search).password;
  const auth = password ? `${username}:${password}@` : '';
  const baseUrl = address ? `http://${address}` : `http://${window.location.host}`;
  const wsUrl = baseUrl.replace("http://", `ws://${auth}`);
  const isNative = qs.parse(window.location.search).mode === "native";

  const onOpen = () => {
    notify("Connected to events", "success");

    if (intervalID) {
      return
    }

    var id = window.setInterval(() => {
      wsRef!.current!.sendMessage(JSON.stringify("keepalive"));
    }, 2000);
    setIntervalID(id);
  };

  const onMessage = (msg: string) => {
    var item: Item = JSON.parse(msg)

    updateData(item, data.Rows);
    setData({ Rows: data.Rows });
  };

  const onClose = () => {
    notify("Event connection lost", "error");

    window.clearInterval(intervalID);
    setIntervalID(0);
  };

  const notify = useCallback((msg: string, variant: VariantType) => {
    props.enqueueSnackbar(msg, {
      variant: variant,
      autoHideDuration: 1000,
      anchorOrigin: {
        vertical: 'bottom',
        horizontal: 'right',
      }
    })
  }, [props]);

  const fetchData = useCallback(() => {
    var headers = new Headers();

    if (password) {
      var authString = `${username}:${password}`
      headers.set('Authorization', 'Basic ' + btoa(authString));
    }

    fetch(`${baseUrl}/?type=json`, { headers: headers }).then(resp => {
      return resp.json().then(data => {
        notify("Data retrieve succesfully", "info");

        setData(data);
      })
    }).catch((e) => {
      notify(`Unable to load or parse topology data ${e}`, "error");

      setTimeout(fetchData, 2000);
    })
  }, [notify, baseUrl]);

  useEffect(() => {
    fetchData();
  }, [props, fetchData, wsRef]);

  return (
    <div className={classes.root}>
      <Websocket ref={wsRef} url={`${wsUrl}/ws`} onOpen={onOpen} onMessage={onMessage} onClose={onClose}
        reconnectIntervalInMilliSeconds={5000} />
      <CssBaseline />
      {!isNative &&
        <React.Fragment>
          <AppBar position="fixed" className={classes.appBar}>
            <Toolbar className={classes.toolbar} variant="dense">
              <Typography component="h1" variant="h6" color="inherit" noWrap className={classes.title}>
                H.A.S.C.
            </Typography>
            </Toolbar>
          </AppBar>
        </React.Fragment>
      }
      <Container className={clsx(classes.container, !isNative && classes.noHeader)}>
        <List className={classes.content}>
          {data.Rows.map((row: Row) => (
            <RenderItem key={row.Item.ID}
              ID={row.Item.ID}
              Type={row.Item.Type}
              Img={row.Item.Img}
              Label={row.Item.Label}
              Value={row.Item.Value}
              Unit={row.Item.Unit}
              LastUpdate={row.Item.LastUpdate}
              SubItems={row.SubItems}
              baseUrl={baseUrl} />
          ))}
        </List>
      </Container>
    </div>
  );
}

export default withSnackbar(App);
