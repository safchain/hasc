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
import { faMemory, faLaptop, faSlidersH, faBug, faLayerGroup, faMicrochip, faBolt } from '@fortawesome/free-solid-svg-icons';
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

interface Props extends WithSnackbarProps { }

interface Item {
  id: string
  oid: string
  type: string
  label: string
  value: string
  img: string
  unit: string
  items?: Array<Item>
  lastupdate: string
}

interface ItemRenderProps {
  className?: string
  oid: string
  type: string
  label: string
  value: string
  img: string
  unit: string
  items?: Array<Item>
  lastupdate: string
  baseUrl: string
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
  }

  return (
    <FontAwesomeIcon icon="chart-bar" />
  )
};

const RenderSwitch: React.FC<ItemRenderProps> = (props) => {
  const onChange = (event: any) => {
    fetch(`${props.baseUrl}/object/` + props.oid, {
      method: 'POST',
      body: event.target.checked ? "ON" : "OFF"
    })
  }

  return (
    <Switch
      edge="end"
      value={props.value === "ON"}
      color="primary"
      onChange={onChange}
    />
  )
};

const RenderButton: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();

  return (
    <Button variant="contained" color="primary" className={props.value === "ON" ? classes.btnGreen : ""}>
      {props.label}
    </Button>
  )
};

const RenderValue: React.FC<ItemRenderProps> = (props) => {
  return (
    <Typography>{props.value}{props.unit}</Typography>
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

  if (props.value === "ON") {
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

  switch (props.type) {
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
      const value = props.value === "OFF" ? "OFF" : "ON";
      const count = (props.value !== "OFF" && props.value !== "ON") ? props.value : "0"

      return (
        <Badge color="primary" badgeContent={count}>
          <Button variant="contained" color="primary" className={props.value === "ON" ? classes.btnGreen : classes.btnRed}>
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

const ItemID = (item: Item): string => {
  if (item.oid && item.id) {
    return item.oid + '-' + item.id
  }
  return item.label.toLowerCase()
};

const RenderItem: React.FC<ItemRenderProps> = (props) => {
  const classes = useStyles();
  const [open, setOpen] = React.useState(false);

  return (
    <React.Fragment>
      <ListItem className={props.className} onClick={() => { setOpen(!open) }}>
        <ListItemIcon>
          <React.Fragment>
            <RenderGroupButton items={props.items} isOpen={open} />
            <RenderIcon img={props.img} />
          </React.Fragment>
        </ListItemIcon>
        <ListItemText primary={props.label} secondary={props.lastupdate ? "Updated: " + props.lastupdate : ""} />
        <ListItemSecondaryAction>
          <RenderType {...props} />
        </ListItemSecondaryAction>
      </ListItem>
      {props.items &&
        <Collapse in={open} timeout="auto" unmountOnExit>
          <List component="div" disablePadding>
            {props.items.map((item: Item) => (
              <RenderItem key={ItemID(item)} className={classes.nested}
                oid={item.oid}
                type={item.type}
                img={item.img}
                label={item.label}
                value={item.value}
                unit={item.unit}
                lastupdate={item.lastupdate}
                items={item.items}
                baseUrl={props.baseUrl}
              />
            ))}
          </List>
        </Collapse>
      }
    </React.Fragment>
  )
};

const updateData = (item: Item, items: Array<Item>) => {
  items.forEach((curr: Item, i: number) => {

    if (curr.id === item.id && curr.oid === item.oid) {
      items[i] = item;
    }

    if (curr.items) {
      updateData(item, curr.items);
    }
  });
};

const App: React.FC<Props> = (props) => {
  const classes = useStyles();
  const [data, setData] = useState({ Items: Array<Item>() });
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

    updateData(item, data.Items);
    setData({ Items: data.Items });
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
          {data.Items.map((item: Item) => (
            <RenderItem key={ItemID(item)}
              oid={item.oid}
              type={item.type}
              img={item.img}
              label={item.label}
              value={item.value}
              unit={item.unit}
              lastupdate={item.lastupdate}
              items={item.items}
              baseUrl={baseUrl} />
          ))}
        </List>
      </Container>
    </div>
  );
}

export default withSnackbar(App);
