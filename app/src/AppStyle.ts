import { makeStyles } from '@material-ui/core/styles';

export const useStyles = makeStyles(theme => ({
  root: {
    height: '100%'
  },
  toolbar: {
    paddingRight: 24, // keep right padding when drawer closed
  },
  toolbarIcon: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
    padding: '0 8px',
    ...theme.mixins.toolbar,
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
  },
  title: {
    flexGrow: 1,
  },
  appBarSpacer: theme.mixins.toolbar,
  content: {
  },
  container: {
    paddingLeft: 0,
    paddingRight: 0,
    paddingBottom: theme.spacing(2),
  },
  noHeader: {
    paddingTop: theme.spacing(5),
  },
  paper: {
    padding: theme.spacing(2),
    display: 'flex',
    overflow: 'auto',
    flexDirection: 'column',
  },
  fabGreen: {
    backgroundColor: "#3fb553"
  },
  fabRed: {
    backgroundColor: "#c74f4f"
  },
  nested: {
    paddingLeft: theme.spacing(5),
  },
  btnGreen: {
    backgroundColor: "#3fb553"
  },
  btnRed: {
    backgroundColor: "#c74f4f"
  }
}));
